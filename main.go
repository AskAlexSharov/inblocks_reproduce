package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"sort"
	"syscall"
	"time"

	"github.com/AskAlexSharov/inblocks_reproduce/mdbx-go"
	"github.com/c2h5oh/datasize"
	"github.com/ledgerwatch/lmdb-go/lmdb"
)

const (
	keysPerBatch    = 1_000
	maxValuesPerKey = 10_000
)

func main() {
	go func() {
		for {
			in, out, _, _ := getRUsage()
			log.Printf("rusage inblocks=%d, outblocks=%d", in, out)
			time.Sleep(5 * time.Second)
		}
	}()
	if len(os.Args) != 3 {
		fmt.Printf(`
use as:
./inblocks_reproduce mdbx write
./inblocks_reproduce mdbx read
./inblocks_reproduce lmdb write
./inblocks_reproduce lmdb reads
`)
		return
	}
	if os.Args[1] == "lmdb" {
		log.Printf("testing lmdb")
		env, dbi := openLmdb()
		defer env.Close()

		if os.Args[1] == "write" {
			writeLmdb(env, dbi)
		} else {
			readLmdb(env, dbi)
		}
		return
	}
	log.Printf("testing mdbx")
	env, dbi := openMdbx()
	defer env.Close()
	if os.Args[1] == "write" {
		writeMdbx(env, dbi)
	} else {
		readMdbx(env, dbi)
	}
}

func openMdbx() (*mdbx.Env, mdbx.DBI) {
	env, err := mdbx.NewEnv()
	if err != nil {
		panic(err)
	}

	if err = env.SetOption(mdbx.OptMaxDB, 100); err != nil {
		panic(err)
	}
	if err = env.SetOption(mdbx.OptMaxReaders, 100); err != nil {
		panic(err)
	}
	if err = env.SetGeometry(-1, -1, int(1*datasize.TB), int(512*datasize.MB), -1, 4*1024); err != nil {
		panic(err)
	}
	if err = env.SetOption(mdbx.OptRpAugmentLimit, 32*1024*1024); err != nil {
		panic(err)
	}

	err = env.Open("./data_mdbx", mdbx.NoReadahead|mdbx.Durable, 0644)
	if err != nil {
		panic(err)
	}
	// 1/8 is good for transactions with a lot of modifications - to reduce invalidation size.
	// But TG app now using Batch and etl.Collectors to avoid writing to DB frequently changing data.
	// It means most of our writes are: APPEND or "single UPSERT per key during transaction"
	if err = env.SetOption(mdbx.OptSpillMinDenominator, 8); err != nil {
		panic(err)
	}
	if err = env.SetOption(mdbx.OptTxnDpInitial, 4*1024); err != nil {
		panic(err)
	}
	if err = env.SetOption(mdbx.OptDpReverseLimit, 4*1024); err != nil {
		panic(err)
	}
	if err = env.SetOption(mdbx.OptTxnDpLimit, 128*1024); err != nil {
		panic(err)
	}
	var dbi mdbx.DBI
	if err := env.Update(func(txn *mdbx.Txn) error {
		txn.RawRead = true
		dbi, err = txn.OpenDBI("alex", mdbx.Create|mdbx.DupSort, nil, nil)
		return err
	}); err != nil {
		panic(err)
	}
	return env, dbi
}

func writeMdbx(env *mdbx.Env, dbi mdbx.DBI) {
	log.Printf("=== insert started")
	for i := 0; i < 100; i++ {
		fileInfo, err := os.Stat("./data_mdbx/mdbx.dat")
		if err != nil {
			panic(err)
		}
		log.Printf("=== insert progress: %d%%, fileSize: %dGb", i, fileInfo.Size()/1024/1024/1024)
		insertBatchMdbx(env, dbi, createBatch(uint8(i)))
	}
}

func readMdbx(env *mdbx.Env, dbi mdbx.DBI) {
	if err := env.View(func(txn *mdbx.Txn) error {
		defer func(t time.Time) { log.Printf("read loop took: %s", time.Since(t)) }(time.Now())
		txn.RawRead = true
		c, err := txn.OpenCursor(dbi)
		if err != nil {
			return err
		}
		for k, _, err := c.Get(nil, nil, mdbx.First); ; k, _, err = c.Get(nil, nil, mdbx.Next) {
			if err != nil {
				if mdbx.IsNotFound(err) {
					break
				}
				return err
			}
			for _, _, err = c.Get(k, nil, mdbx.FirstDup); ; _, _, err = c.Get(k, nil, mdbx.NextDup) {
				if err != nil {
					if mdbx.IsNotFound(err) {
						break
					}
					return err
				}
			}
		}
		return nil
	}); err != nil {
		panic(err)
	}
}
func createBatch(batchId uint8) []*Pair {
	val := make([]byte, 44)
	key := make([]byte, 20)
	key[0] = batchId
	var pairs []*Pair
	for j := 0; j < keysPerBatch; j++ {
		//_, err := rand.Read(key)
		//if err != nil {
		//	panic(err)
		//}
		key = next(key)
		for h := 0; h < maxValuesPerKey; h++ {
			val = next(val)
			pairs = append(pairs, &Pair{k: copyBytes(key), v: copyBytes(val)})
		}

	}
	//sortPairs(pairs)

	return pairs
}

func openLmdb() (*lmdb.Env, lmdb.DBI) {
	env, err := lmdb.NewEnv()
	if err != nil {
		panic(err)
	}

	if err = env.SetMaxDBs(100); err != nil {
		panic(err)
	}
	if err = env.SetMapSize(int64(1 * datasize.TB)); err != nil {
		panic(err)
	}
	if err = os.MkdirAll("./data_lmdb", 0744); err != nil {
		panic(err)
	}
	err = env.Open("./data_lmdb", mdbx.NoReadahead, 0644)
	if err != nil {
		panic(err)
	}
	var dbi lmdb.DBI
	if err := env.Update(func(txn *lmdb.Txn) error {
		txn.RawRead = true
		dbi, err = txn.OpenDBI("alex", lmdb.Create|lmdb.DupSort)
		return err
	}); err != nil {
		panic(err)
	}

	return env, dbi
}
func writeLmdb(env *lmdb.Env, dbi lmdb.DBI) {
	log.Printf("=== in sert started")
	for i := 0; i < 100; i++ {
		fileInfo, err := os.Stat("./data_lmdb/data.db")
		if err != nil {
			panic(err)
		}
		log.Printf("=== insert progress: %d%%, fileSize: %dGb", i, fileInfo.Size()/1024/1024/1024)
		insertBatchLmdb(env, dbi, createBatch(uint8(i)))
	}
}

func readLmdb(env *lmdb.Env, dbi lmdb.DBI) {
	for i := 0; i < 10; i++ {
		if err := env.View(func(txn *lmdb.Txn) error {
			defer func(t time.Time) { log.Printf("read loop took: %s", time.Since(t)) }(time.Now())
			txn.RawRead = true
			c, err := txn.OpenCursor(dbi)
			if err != nil {
				return err
			}
			for _, _, err = c.Get(nil, nil, lmdb.First); ; _, _, err = c.Get(nil, nil, lmdb.Next) {
				if err != nil {
					if lmdb.IsNotFound(err) {
						break
					}
					return err
				}
				for _, _, err = c.Get(nil, nil, lmdb.FirstDup); ; _, _, err = c.Get(nil, nil, lmdb.NextDup) {
					if err != nil {
						if lmdb.IsNotFound(err) {
							break
						}
						return err
					}
				}
			}
			return err
		}); err != nil {
			panic(err)
		}
	}
}

func insertBatchMdbx(env *mdbx.Env, dbi mdbx.DBI, pairs []*Pair) {
	if err := env.Update(func(txn *mdbx.Txn) error {
		txn.RawRead = true
		c, err := txn.OpenCursor(dbi)
		if err != nil {
			return err
		}
		defer c.Close()

		for _, pair := range pairs {
			k, v := pair.k, pair.v
			err = c.Put(k, v, mdbx.AppendDup)

			//_, _, err := c.Get(k, v, mdbx.GetBoth)
			//if err != nil {
			//	if mdbx.IsNotFound(err) {
			//		err = c.Put(k, v, mdbx.Upsert)
			//		if err != nil {
			//			panic(err)
			//		}
			//		continue
			//	}
			//	panic(err)
			//}
			//err = c.Del(mdbx.Current)
			//if err != nil {
			//	panic(err)
			//}
			//
			//err = c.Put(k, v, mdbx.Upsert)
			if err != nil {
				panic(err)
			}
		}

		return nil
	}); err != nil {
		panic(err)
	}
}

func insertBatchLmdb(env *lmdb.Env, dbi lmdb.DBI, pairs []*Pair) {
	if err := env.Update(func(txn *lmdb.Txn) error {
		txn.RawRead = true
		c, err := txn.OpenCursor(dbi)
		if err != nil {
			return err
		}
		defer c.Close()

		for _, pair := range pairs {
			k, v := pair.k, pair.v
			err = c.Put(k, v, lmdb.AppendDup)

			//_, _, err := c.Get(k, v, lmdb.GetBoth)
			//if err != nil {
			//	if lmdb.IsNotFound(err) {
			//		err = c.Put(k, v, 0)
			//		if err != nil {
			//			panic(err)
			//		}
			//		continue
			//	}
			//	panic(err)
			//}
			//err = c.Del(lmdb.Current)
			//if err != nil {
			//	panic(err)
			//}
			//
			//err = c.Put(k, v, 0)
			if err != nil {
				panic(err)
			}
		}

		return nil
	}); err != nil {
		panic(err)
	}
}

func sortPairs(pairs []*Pair) {
	sort.Slice(pairs, func(i, j int) bool {
		cmp := bytes.Compare(pairs[i].k, pairs[j].k)
		if cmp == 0 {
			cmp = bytes.Compare(pairs[i].v, pairs[j].v)
		}
		return cmp < 0
	})
}

type Pair struct {
	k []byte
	v []byte
}

func getRUsage() (inBlock, outBlocks, nvcsw, nivcsw int64) {
	var ru syscall.Rusage
	if err := syscall.Getrusage(syscall.RUSAGE_SELF, &ru); err != nil {
		log.Fatal("Failed to retrieve CPU time", "err", err)
		return
	}
	return ru.Inblock, ru.Oublock, ru.Nvcsw, ru.Nivcsw
}

// next does []byte++
func next(in []byte) []byte {
	r := make([]byte, len(in))
	copy(r, in)
	for i := len(r) - 1; i >= 0; i-- {
		if r[i] != 255 {
			r[i]++
			return r
		}

		r[i] = 0
	}
	return r
}

func copyBytes(b []byte) (copiedBytes []byte) {
	if b == nil {
		return nil
	}
	copiedBytes = make([]byte, len(b))
	copy(copiedBytes, b)

	return
}
