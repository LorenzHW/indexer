package test

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"runtime"

	"github.com/algorand/go-algorand/data/basics"
	"github.com/algorand/go-algorand/data/transactions"
	"github.com/algorand/go-algorand/protocol"

	"github.com/algorand/indexer/idb"
	"github.com/algorand/indexer/util"
)

var quiet = false
var exitValue = 0

// SetQuiet quiet mode of this logging thing.
func SetQuiet(q bool) {
	quiet = q
}

// ExitValue returns the captured exit value.
func ExitValue() int {
	return exitValue
}

func info(format string, a ...interface{}) {
	if quiet {
		return
	}
	fmt.Printf(format, a...)
}

// Info is the the only logging level for this thing.
var Info = info

func infoln(s string) {
	if quiet {
		return
	}
	fmt.Println(s)
}

func myStackTrace() {
	for skip := 1; skip < 3; skip++ {
		_, file, line, ok := runtime.Caller(skip)
		if !ok {
			return
		}
		fmt.Fprintf(os.Stderr, "%s:%d\n", file, line)
	}
}

// PrintAssetQuery prints information about an asset query.
func PrintAssetQuery(db idb.IndexerDb, q idb.AssetsQuery) {
	count := uint64(0)
	assetchan, _ := db.Assets(context.Background(), q)
	for ar := range assetchan {
		util.MaybeFail(ar.Error, "asset query %v\n", ar.Error)
		pjs, err := json.Marshal(ar.Params)
		util.MaybeFail(err, "json.Marshal params %v\n", err)
		var creator basics.Address
		copy(creator[:], ar.Creator)
		info("%d %s %s\n", ar.AssetID, creator.String(), pjs)
		count++
	}
	info("%d rows\n", count)
	if q.Limit != 0 && q.Limit != count {
		fmt.Fprintf(os.Stderr, "asset q CAME UP SHORT, limit=%d actual=%d, q=%#v\n", q.Limit, count, q)
		myStackTrace()
		exitValue = 1
	}
}

// PrintAccountQuery prints information about an account query.
func PrintAccountQuery(db idb.IndexerDb, q idb.AccountQueryOptions) {
	accountchan, _ := db.GetAccounts(context.Background(), q)
	count := uint64(0)
	for ar := range accountchan {
		util.MaybeFail(ar.Error, "GetAccounts err %v\n", ar.Error)
		jb, err := json.Marshal(ar.Account)
		util.MaybeFail(err, "err %v\n", err)
		infoln(string(jb))
		//fmt.Printf("%#v\n", ar.Account)
		count++
	}
	info("%d accounts\n", count)
	if q.Limit != 0 && q.Limit != count {
		fmt.Fprintf(os.Stderr, "account q CAME UP SHORT, limit=%d actual=%d, q=%#v\n", q.Limit, count, q)
		myStackTrace()
		exitValue = 1
	}
}

// PrintTxnQuery prints information about a transaction query.
func PrintTxnQuery(db idb.IndexerDb, q idb.TransactionFilter) {
	rowchan, _ := db.Transactions(context.Background(), q)
	count := uint64(0)
	for txnrow := range rowchan {
		util.MaybeFail(txnrow.Error, "err %v\n", txnrow.Error)
		var stxn transactions.SignedTxnWithAD
		err := protocol.Decode(txnrow.TxnBytes, &stxn)
		util.MaybeFail(err, "decode txnbytes %v\n", err)
		tjs := util.JSONOneLine(stxn.Txn)
		info("%d:%d %s sr=%d rr=%d ca=%d cr=%d t=%s\n", txnrow.Round, txnrow.Intra, tjs, stxn.SenderRewards, stxn.ReceiverRewards, stxn.ClosingAmount, stxn.CloseRewards, txnrow.RoundTime.String())
		count++
	}
	info("%d txns\n", count)
	if q.Limit != 0 && q.Limit != count {
		fmt.Fprintf(os.Stderr, "txn q CAME UP SHORT, limit=%d actual=%d, q=%#v\n", q.Limit, count, q)
		myStackTrace()
		exitValue = 1
	}
}
