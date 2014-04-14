//convert-rf-readings is a small app to switch between <band> referenced readings and <bandless> reference readings
//within the HouseMon app (https://github.com/jcw/housemon) (v0.9.0)
//if you want to use <band> referenced readings you can use this package: https://github.com/TheDistractor/flow-ext/tree/master/gadgets/housemon/rf
//until/unless the core system has this support added
package main

import (
	"encoding/json"
	"fmt"
	"github.com/codegangsta/cli"
	"github.com/syndtr/goleveldb/leveldb"
	iterator "github.com/syndtr/goleveldb/leveldb/iterator"
	opt "github.com/syndtr/goleveldb/leveldb/opt"
	dbutil "github.com/syndtr/goleveldb/leveldb/util"
	"os"
	"path"
	"strings"
)


type Reading struct {
	Ms  int64          `json:"ms"`
	Val map[string]int `json:"val"`
	Loc string         `json:"loc"`
	Typ string         `json:"typ"`
	Id  string         `json:"id"`
}

func main() {

	app := cli.NewApp()
	app.Name = "convert-rf-readings"
	app.Version = "0.9.0"
	app.Usage = "converts your HouseMon database readings from:\n\tRF:<group>:<node>\n\tto\n\tRF:<band>:<group>:<node>"
	app.Action = func(c *cli.Context) {
		convertTo(c)
	}
	app.Flags = []cli.Flag{
		cli.StringFlag{Name: "band", Value: "868", Usage: "the default band for conversion (e.g one of 433,868,915)"},
		cli.StringFlag{Name: "path", Value: "./data", Usage: "the default path to your existing database"},
	}
	app.Commands = []cli.Command{
		{
			Name:        "revert",
			ShortName:   "r",
			Usage:       "revert",
			Description: "reverts a conversion by removing <band>",
			Action: func(c *cli.Context) {
				convertFrom(c)
			},
		},
		{
			Name:        "list",
			ShortName:   "l",
			Usage:       "list",
			Description: "lists existing reading keys",
			Action: func(c *cli.Context) {
				list(c)
			},
		},

	}
	app.Run(os.Args)

}

func convertFrom(c *cli.Context) {
	success := true
	defaultBand := c.GlobalString("band")
	dbPath := c.GlobalString("path")

	pwd, err := os.Getwd()
	dbPath = path.Join(pwd, dbPath)

	options := &opt.Options{ErrorIfMissing: true}

	db, err := leveldb.OpenFile(dbPath, options)

	if err != nil {
		msg := fmt.Sprintf("%s %s", err, dbPath)
		panic(msg)
	}

	fmt.Println("using database:", dbPath)

	defer db.Close()

	var iter iterator.Iterator

	fmt.Println("Converting FROM band format...")

	iter = db.NewIterator(&dbutil.Range{Start: []byte("/reading/RF12:"), Limit: []byte("/reading/RF12~")}, nil)
	for iter.Next() {
		key := string(iter.Key())
		rval := iter.Value()

		kparts := strings.Split(key, "/")

		rfnet := kparts[len(kparts)-1:]
		rfparts := strings.Split(rfnet[0], ":")

		if len(rfparts) == 4 { //its an new format with band

			if rfparts[1] == defaultBand {
				fmt.Println("INFO:Downgrading:", key)

				//remove band
				copy(rfparts[1:], rfparts[1+1:])
				rfparts[len(rfparts)-1] = ""
				rfparts = rfparts[:len(rfparts)-1]

				//adjust the json structure
				var reading Reading
				err := json.Unmarshal(rval, &reading)
				if err == nil {
					id := fmt.Sprintf("%s", strings.Join(rfparts, ":"))
					reading.Id = id
					fmt.Println("INFO:New reading:", reading)
					data, err := json.Marshal(reading)
					if err == nil {

						//write a new key
						newkey := "/reading/" + id

						err = db.Put([]byte(newkey), []byte(data), nil)
						if err != nil {
							fmt.Println("ERR:Write failed for:", newkey)
							success = false
							continue
						}

						fmt.Println("INFO:New Reading Stored:", newkey)

						//remove old key
						err = db.Delete([]byte(key), nil)

						if err != nil {
							success = false
							fmt.Println("WARN: could not remove old key:", key)
						}

					}

				} else {
					fmt.Println("Err:Failed to unmarshal data:", err)
					success = false
				}

			}

		}

	}
	iter.Release()
	err = iter.Error()
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("Downgrade finished")
	if !success {
		fmt.Println("problems detected, please review screen")
	}

}

func list(c *cli.Context) {

	dbPath := c.GlobalString("path")

	pwd, err := os.Getwd()
	dbPath = path.Join(pwd, dbPath)

	options := &opt.Options{ErrorIfMissing: true}

	db, err := leveldb.OpenFile(dbPath, options)

	if err != nil {
		msg := fmt.Sprintf("%s %s", err, dbPath)
		panic(msg)
	}

	fmt.Println("using database:", dbPath)

	defer db.Close()

	var iter iterator.Iterator

	fmt.Println("Listing reading keys...")

	iter = db.NewIterator(&dbutil.Range{Start: []byte("/reading/RF12:"), Limit: []byte("/reading/RF12~")}, nil)
		for iter.Next() {
			key := string(iter.Key())
			fmt.Println(key)
		}
		iter.Release()

	//err = iter.Error()

}

func convertTo(c *cli.Context) {

	success := true
	defaultBand := c.String("band")
	dbPath := c.String("path")

	pwd, err := os.Getwd()
	dbPath = path.Join(pwd, dbPath)

	options := &opt.Options{ErrorIfMissing: true}

	db, err := leveldb.OpenFile(dbPath, options)

	if err != nil {
		msg := fmt.Sprintf("%s %s", err, dbPath)
		panic(msg)
	}

	fmt.Println("using database:", dbPath)

	defer db.Close()

	var iter iterator.Iterator

	fmt.Println("Converting TO band format...")

	iter = db.NewIterator(&dbutil.Range{Start: []byte("/reading/RF12:"), Limit: []byte("/reading/RF12~")}, nil)
	for iter.Next() {

		key := string(iter.Key())
		rval := iter.Value()

		kparts := strings.Split(key, "/")

		rfnet := kparts[len(kparts)-1:]
		rfparts := strings.Split(rfnet[0], ":")

		if len(rfparts) == 3 { //its an original without band

			fmt.Println("INFO:Upgrading:", key)

			//insert band
			rfparts = append(rfparts, "")
			copy(rfparts[1+1:], rfparts[1:])
			rfparts[1] = defaultBand

			//adjust the json structure
			var reading Reading
			err := json.Unmarshal(rval, &reading)
			if err == nil {
				id := fmt.Sprintf("%s", strings.Join(rfparts, ":"))
				reading.Id = id
				fmt.Println("INFO:New reading:", reading)
				data, err := json.Marshal(reading)
				if err == nil {

					//write a new key
					newkey := "/reading/" + id

					err = db.Put([]byte(newkey), []byte(data), nil)
					if err != nil {
						fmt.Println("ERR:Write failed for:", newkey)
						success = false
						continue
					}

					fmt.Println("INFO:New Reading Stored:", newkey)

					//remove old key
					err = db.Delete([]byte(key), nil)

					if err != nil {
						success = false
						fmt.Println("WARN: could not remove old key:", key)
					}

				}

			} else {
				fmt.Println("Err:Failed to unmarshal data:", err)
				success = false
			}

		}

	}
	iter.Release()
	err = iter.Error()
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("Upgrade finished")
	if !success {
		fmt.Println("problems detected, please review screen")
	}

}
