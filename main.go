package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"

	dbm "github.com/cometbft/cometbft-db"
	"github.com/cosmos/cosmos-sdk/store/iavl"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	gogotypes "github.com/cosmos/gogoproto/types"
)

func main() {

	flag.Parse()
	args := flag.Args()

	// usage: dump <dataDir> <height(optional)>
	if len(args) < 1 {
		fmt.Println("usage: dump <dataDir> <height(optional)>")
		return
	}

	rootDir := args[0]
	targetHeight := int64(-1)
	if len(args) > 1 {
		targetHeight, _ = strconv.ParseInt(args[1], 10, 64)
	}

	fmt.Println("dir:", rootDir)

	dataDir := filepath.Join(rootDir, "data")

	// Check directory exists
	if _, err := os.Stat(dataDir); os.IsNotExist(err) {
		fmt.Println("dataDir not found")
		return
	}

	// Open DB
	db, err := dbm.NewDB("application", dbm.GoLevelDBBackend, dataDir)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	// Get latest block height
	bz, err := db.Get([]byte("s/latest"))
	if err != nil {
		panic(err)
	}

	var latestHeight int64
	err = gogotypes.StdInt64Unmarshal(&latestHeight, bz)
	if err != nil {
		panic(err)
	}

	if targetHeight == -1 {
		targetHeight = latestHeight
	}

	// Print info
	fmt.Println("latestHeight: ", latestHeight)
	fmt.Println("targetHeight: ", targetHeight)

	// s/<height> is CommitInfo
	bz, err = db.Get([]byte(fmt.Sprintf("s/%d", targetHeight)))
	if err != nil {
		panic(err)
	}

	cInfo := storetypes.CommitInfo{}
	err = cInfo.Unmarshal(bz)
	if err != nil {
		panic(err)
	}

	// transorm array to map
	infos := make(map[string]storetypes.StoreInfo)
	for _, info := range cInfo.StoreInfos {
		infos[info.Name] = info
	}

	modules := make([]string, 0, len(infos))
	for module := range infos {
		modules = append(modules, module)
	}

	sort.Strings(modules)

	// Open DB as PrefixDB
	stores := make(map[string]storetypes.CommitKVStore)
	for _, module := range modules {

		targetInfo := infos[module]

		prefix := "s/k:" + module + "/"
		prefixDB := dbm.NewPrefixDB(db, []byte(prefix))

		var id storetypes.CommitID = targetInfo.CommitId
		targetStore, err := iavl.LoadStore(prefixDB, nil, nil, id, false, 0, false)
		if err != nil {
			panic(err)
		}

		stores[module] = targetStore
	}

	fmt.Println()

	// Print key-value
	for _, module := range modules {
		targetStore := stores[module]

		itr := targetStore.Iterator(nil, nil)
		for ; itr.Valid(); itr.Next() {
			if isASCII(itr.Key()) {
				fmt.Printf("key(ascii): %s %s\n", module, string(itr.Key()))
			} else {
				fmt.Printf("key(hex): %s %x\n", module, itr.Key())
			}
			fmt.Printf("value(hex): %x\n", itr.Value())
			fmt.Println()
		}
	}
}

func isASCII(b []byte) bool {
	for _, c := range b {
		if c > 127 || c < 32 {
			return false
		}
	}
	return true
}
