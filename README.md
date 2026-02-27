# Distributed Locks in MongoDB

[![Go Report Card](https://goreportcard.com/badge/github.com/foomo/mongo-lock)](https://goreportcard.com/report/github.com/foomo/mongo-lock)
[![GoDoc](https://godoc.org/github.com/foomo/mongo-lock?status.svg)](https://godoc.org/github.com/foomo/mongo-lock)

> This is a fork of the wonderful [github.com/square/mongo-lock](https://github.com/square/mongo-lock) by Square, Inc.
> The original repository is no longer actively maintained, so this fork continues development under the same [Apache 2.0 license](LICENSE).

This package provides a Go client for creating distributed locks in MongoDB.

## Setup
Install the package with "go get".
```
go get "github.com/foomo/mongo-lock"
```

In order to use it, you must have an instance of MongoDB running with a collection that can be used to store locks.
All of the examples here will assume the collection name is "locks", but you can change it to whatever you want.

#### Required Indexes
There is one index that is required in order for this package to work:
```
db.locks.createIndex( { resource: 1 }, { unique: true } )
```

#### Recommended Indexes
The following indexes are recommend to help the performance of certain queries:
```
db.locks.createIndex( { "exclusive.lockId": 1 } )
db.locks.createIndex( { "exclusive.expiresAt": 1 } )
db.locks.createIndex( { "shared.locks.lockId": 1 } )
db.locks.createIndex( { "shared.locks.expiresAt": 1 } )
```

The [Client.CreateIndexes](https://godoc.org/github.com/foomo/mongo-lock#Client.CreateIndexes) method can be called to create all of the required and recommended indexes.

#### Recommended Write Concern
To minimize the risk of losing locks when one or more nodes in your replica set fail, setting the write acknowledgement for the session to "majority" is recommended.

## Usage
Here is an example of how to use this package:
```go
package main

import (
    "context"
    "log"
    "time"

    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
    "go.mongodb.org/mongo-driver/mongo/writeconcern"

    "github.com/foomo/mongo-lock"
)

func main() {
    // Create a Mongo session and set the write mode to "majority".
    mongoUrl := "youMustProvideThis"
    database := "dbName"
    collection := "collectionName"

    ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
    defer cancel()

    m, err := mongo.Connect(ctx, options.Client().
        ApplyURI(mongoUrl).
        SetWriteConcern(writeconcern.New(writeconcern.WMajority())))

    if err != nil {
        log.Fatal(err)
    }

    defer func() {
        if err = m.Disconnect(ctx); err != nil {
            panic(err)
        }
    }()

    // Configure the client for the database and collection the lock will go into.
    col := m.Database(database).Collection(collection)

    // Create a MongoDB lock client.
    c := lock.NewClient(col)

    // Create the required and recommended indexes.
    c.CreateIndexes(ctx)

    lockId := "abcd1234"

    // Create an exclusive lock on resource1.
    err = c.XLock(ctx, "resource1", lockId, lock.LockDetails{})
    if err != nil {
        log.Fatal(err)
    }

    // Create a shared lock on resource2.
    err = c.SLock(ctx, "resource2", lockId, lock.LockDetails{}, -1)
    if err != nil {
        log.Fatal(err)
    }

    // Unlock all locks that have our lockId.
    _, err = c.Unlock(ctx, lockId)
    if err != nil {
        log.Fatal(err)
    }
}


```

## How It Works
This package can be used to create both shared and exclusive locks.
To create a lock, all you need is the name of a resource (the object that gets locked) and a lockId (lock identifier).
Multiple locks can be created with the same lockId, which makes it easy to unlock or renew a group of related locks at the same time.
Another reason for using lockIds is to ensure the only the client that creates a lock knows the lockId needed to unlock it (i.e., knowing a resource name alone is not enough to unlock it).

Here is a list of rules that the locking behavior follows
* A resource can only have one exclusive lock on it at a time.
* A resource can have multiple shared locks on it at a time [1][2].
* A resource cannot have both an exclusive lock and a shared lock on it at the same time.
* A resource can have no locks on it at all.

[1] It is possible to limit the number of shared locks that can be on a resource at a time (see the docs for [Client.SLock](https://godoc.org/github.com/foomo/mongo-lock#Client.SLock) for more details).
[2] A resource can't have more than one shared lock on it with the same lockId at a time.

#### Additional Features
* **TTLs**: You can optionally set a time to live (TTL) when creating a lock. If you do not set one, the lock will not have a TTL. TTLs can be renewed via the [Client.Renew](https://godoc.org/github.com/foomo/mongo-lock#Client.Renew) method as long as all of the locks associated with a given lockId have a TTL of at least 1 second (or no TTL at all). There is no automatic process to clean up locks that have outlived their TTL, but this package does provide a [Purger](https://godoc.org/github.com/foomo/mongo-lock#Purger) that can be run in a loop to accomplish this.


## Schema
Resources are the only documents stored in MongoDB. Locks on a resource are stored within the resource documents, like so:
```
{
        "resource" : "resource1",
        "exclusive" : {
                "lockId" : null,
                "owner" : null,
                "host" : null,
                "createdAt" : null,
                "renewedAt" : null,
                "expiresAt" : null,
                "acquired" : false
        },
        "shared" : {
                "count" : 1,
                "locks" : [
                        {
                                "lockId" : "abcd",
                                "owner" : "john",
                                "host" : "host.name",
                                "createdAt" : ISODate("2018-01-25T01:58:47.243Z"),
                                "renewedAt" : null,
                                "expiresAt" : null,
                                "acquired" : true,
                        }
                ]
        }
}
```
Note: shared locks are stored as an array instead of a map (keyed on lockId) so that shared lock fields can be indexed.
This helps with the performance of unlocking, renewing, and getting the status of locks.

## Development

#### Prerequisites
This project uses [mise](https://mise.jdx.dev/) to manage tool versions (golangci-lint, lefthook). Install it and run:
```
mise install
```

#### Commands
Run `make help` to see all available targets. Key targets:
```
make check      # Full pipeline: tidy, generate, lint, test
make lint       # Run golangci-lint
make lint.fix   # Auto-fix lint violations
make test       # Run tests with coverage
make test.race  # Run tests with race detection
```
## How to Contribute

Contributions are welcome! Please read the [contributing guide](docs/CONTRIBUTING.md).

![Contributors](https://contributors-table.vercel.app/image?repo=foomo/mongo-lock&width=50&columns=15)

## License
Copyright 2018 Square, Inc.
Copyright 2026 foomo by bestbytes

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

  http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
