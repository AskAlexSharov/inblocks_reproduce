package main

import (
	"bytes"
	"crypto/rand"
	mathRand "math/rand"
	"sort"

	"github.com/AskAlexSharov/inblocks_reproduce/mdbx-go"
)

func main() {
	env, err := mdbx.NewEnv()
	if err != nil {
		panic(err)
	}
	err = env.Open("./data", mdbx.WriteMap|mdbx.NoReadahead|mdbx.Durable, 0644)
	if err != nil {
		panic(err)
	}
	defer env.Close()

	var dbi mdbx.DBI
	if err = env.Update(func(txn *mdbx.Txn) error {
		txn.RawRead = true
		dbi, err = txn.OpenDBI("alex", mdbx.Create|mdbx.DupSort, nil, nil)
		return err
	}); err != nil {
		panic(err)
	}

	for i := 0; i < 100; i++ {
		insertBatch(env, dbi)
	}
}

func insertBatch(env *mdbx.Env, dbi mdbx.DBI) {
	keysPerBatch := 100
	maxValuesPerKey := 100
	var pairs []*Pair
	for j := 0; j < keysPerBatch; j++ {
		key := make([]byte, 20)
		_, err := rand.Read(key)
		if err != nil {
			panic(err)
		}
		for h := 0; h < mathRand.Intn(maxValuesPerKey); h++ {
			pairs = append(pairs, NewPair(key))
		}

	}
	sortPairs(pairs)

	if err := env.Update(func(txn *mdbx.Txn) error {
		txn.RawRead = true
		c, err := txn.OpenCursor(dbi)
		if err != nil {
			return err
		}
		defer c.Close()

		for _, pair := range pairs {
			k, v := pair.k, pair.v
			_, _, err := c.Get(k, v, mdbx.GetBoth)
			if err != nil {
				if mdbx.IsNotFound(err) {
					err = c.Put(k, v, mdbx.Upsert)
					if err != nil {
						panic(err)
					}
				}
				panic(err)
			}
			err = c.Del(mdbx.Current)
			if err != nil {
				panic(err)
			}

			err = c.Put(k, v, mdbx.Upsert)
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

func NewPair(k []byte) *Pair {
	v := make([]byte, 44)
	_, err := rand.Read(v)
	if err != nil {
		panic(err)
	}
	return &Pair{k: k, v: v}
}
