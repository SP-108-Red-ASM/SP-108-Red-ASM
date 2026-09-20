package main

import (
	"encoding/json"
	"fmt"
	"os"
)

func main() {
	err := run()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	c, err := cfg()
	if err != nil {
		return err
	}

	//db, err := open(c.DBPath)
	//if err != nil {
	//	return err
	//}
	//defer db.Close()

	//cyc, err := newCycle(db)
	//if err != nil {
	//	return err
	//}

	hosts, err := scan(c)
	if err != nil {
		return err
	}

	b, err := json.MarshalIndent(hosts, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(b))

	//chgs, err := diff(db, cyc)
	//if err != nil {
	//	return err
	//}
	//
	//err = report(db, cyc, chgs)
	//if err != nil {
	//	return err
	//}

	return nil
}
