package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	dbm "github.com/cometbft/cometbft-db"
	"github.com/cosmos/cosmos-sdk/store"
	"github.com/cosmos/cosmos-sdk/store/iavl"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
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

	// Get modules from DB keys
	_modules := make(map[string]bool)
	itr, err := db.Iterator(nil, nil)
	if err != nil {
		panic(err)
	}

	for ; itr.Valid(); itr.Next() {
		key := string(itr.Key())
		split := strings.Split(key, "/")
		if len(split) < 2 {
			continue
		}
		if split[0] != "s" {
			continue
		}
		if len(split[1]) < 2 {
			continue
		}
		if split[1][0] != 'k' {
			continue
		}
		module := split[1][2:]
		_modules[module] = true
	}

	modules := make([]string, 0)
	for module := range _modules {
		modules = append(modules, module)
	}
	// go map order is random, so sort to diff
	sort.Strings(modules)

	// Open Multistore
	ms := store.NewCommitMultiStore(db)
	for _, module := range modules {
		ms.MountStoreWithDB(storetypes.NewKVStoreKey(module), storetypes.StoreTypeIAVL, db)
	}

	if targetHeight == -1 {
		targetHeight = ms.LatestVersion()
	}

	// Print info
	fmt.Println("latestVersion: ", ms.LatestVersion())
	fmt.Println("targetHeight: ", targetHeight)
	fmt.Println("modules:")
	for _, module := range modules {
		fmt.Println("  ", module)
	}

	// s/<height> is CommitInfo
	bz, err := db.Get([]byte(fmt.Sprintf("s/%d", targetHeight)))
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

	// Open DB as PrefixDB
	stores := make(map[string]storetypes.CommitKVStore)
	for _, module := range modules {

		targetInfo := infos[module]

		prefix := "s/k:" + module + "/"
		prefixDB := dbm.NewPrefixDB(db, []byte(prefix))

		var storekey storetypes.StoreKey = storetypes.NewKVStoreKey("*") // FIXME: I don't know this argument means
		var id storetypes.CommitID = targetInfo.CommitId
		targetStore, err := iavl.LoadStore(prefixDB, nil, storekey, id, false, 1000, false) // there are no particular reason to set 1000
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
