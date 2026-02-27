// Copyright 2018, Square, Inc.

package lock_test

import (
	"context"
	"sort"
	"testing"
	"time"

	lock "github.com/foomo/mongo-lock"
	"github.com/stretchr/testify/require"
	"gotest.tools/v3/assert"
)

func TestPurge(t *testing.T) {
	// setup and teardown are defined in lock_test.go
	collection := setup(t)

	ctx, cancel := context.WithTimeout(t.Context(), time.Second*60)
	defer cancel()

	client := lock.NewClient(collection)

	// Create some locks.
	err := client.XLock(ctx, "resource1", "aaaa", lock.LockDetails{})
	require.NoError(t, err)

	err = client.XLock(ctx, "resource2", "bbbb", lock.LockDetails{TTL: 1})
	require.NoError(t, err)

	err = client.SLock(ctx, "resource3", "cccc", lock.LockDetails{TTL: 1}, -1)
	require.NoError(t, err)

	// Sleep for a second to let TTLs expire
	time.Sleep(time.Duration(1500) * time.Millisecond)

	// Purge the locks.
	purger := lock.NewPurger(client)

	purged, err := purger.Purge(ctx)
	require.NoError(t, err)

	if len(purged) != 2 {
		t.Errorf("%d locks purged, expected %d", len(purged), 2)
	}

	var purgedSorted lock.LockStatusesByCreatedAtDesc = purged
	sort.Sort(purgedSorted)

	assert.Equal(t, "resource3", purged[0].Resource)
	assert.Equal(t, "resource2", purged[1].Resource)
}

func TestPurgeSameLockIdDiffTTLs(t *testing.T) {
	// setup and teardown are defined in lock_test.go
	collection := setup(t)

	ctx, cancel := context.WithTimeout(t.Context(), time.Second*60)
	defer cancel()

	client := lock.NewClient(collection)

	// Create some locks with different TTLs, all owned by the same lockId.
	err := client.XLock(ctx, "resource1", "aaaa", lock.LockDetails{}) // no TTL
	require.NoError(t, err)

	err = client.XLock(ctx, "resource2", "aaaa", lock.LockDetails{TTL: 30})
	require.NoError(t, err)

	err = client.SLock(ctx, "resource3", "aaaa", lock.LockDetails{TTL: 1}, -1)
	require.NoError(t, err)

	// Sleep for a second to let some TTLs expire
	time.Sleep(time.Duration(1500) * time.Millisecond)

	// Purge the locks.
	purger := lock.NewPurger(client)

	purged, err := purger.Purge(ctx)
	require.NoError(t, err)

	assert.Equal(t, 3, len(purged))

	var purgedSorted lock.LockStatusesByCreatedAtDesc = purged
	sort.Sort(purgedSorted)

	assert.Equal(t, "resource3", purged[0].Resource)
	assert.Equal(t, "resource2", purged[1].Resource)
	assert.Equal(t, "resource1", purged[2].Resource)
}
