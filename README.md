## Turing
Upload the data to the arweave network in the specified order and download it from the arweave network in the order of upload.

### Implements
turing implements the above functionality through the rollup & tracker service

### rollup server
rollup is responsible for uploading the data. The data is uploaded in strict incoming order and is guaranteed to be 100% available.

Example:
```go
// New rollup
suggestLastArTxId := "" 	        // If the service is restarted in the case of a lost bolt.db, you need to specify the last txId of the rollup owner uplink
arNode := "https://arweave.net"
arWalletKeyPath := "./key.json"   // rollup owner arweave key store
owner := "" 	// rollup owner
tags := []{} 	// rollup on chain tx tags
dbPath := ""  // bolt.db store file path,default value is "./bolt"
rollup := New(suggestLastArTxId, arNode, arWalletKeyPath, owner, tags, dbPath)

// run server
timeInterval := 2 * time.Minute  // rollup frequency
maxOfRollup := 99999 // maximum number of transactions per rollup
rollup.Run(timeInterval, maxOfRollup)

// add data
data := []byte("****************")
rollup.AddTx <- data
```

### tracker server
The tracker listens to the data uploaded by the rollup, downloads the data locally and has the ability to filter the duplicate data uploaded by the rollup.

Example:
```go
// New tracker
tags := []{}			       // need tracker tx tags
arNode := "https://arweave.net"
arOwner := ""           // always the same as rollup owner
dbPath := ""            // bolt.db store file path,default value is "./bolt"
cursor := 0             // bolt.db cursor, options always is 0
tracker := New(tags, arNode, arOwner, dbPath)

// run tracker
trcker.Run(cursor)

// channal to get 
data := <- tracker.SubscribeTx()
```
