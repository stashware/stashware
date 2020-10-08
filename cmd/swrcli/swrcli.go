package main

import (
	"os"

	"stashware/oo"

	"github.com/urfave/cli/v2"
)

func main() {
	local := []*cli.Command{
		info_cmd,
		peers_cmd,
		tx_pending_cmd,
		get_block_cmd,
		tx_by_id_cmd,
		tx_data_cmd,
		price_cmd,
		wallet_new_cmd,
		wallet_by_addr_cmd,
		submit_block_cmd,
		submit_tx_cmd,
		get_wallet_list_cmd,
		get_addr_txs_cmd,
	}

	app := &cli.App{
		Name:  "swr client",
		Usage: "A swr client application",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "host",
				Value: "127.0.0.1",
			},
			&cli.Int64Flag{
				Name:  "port",
				Value: 3080,
			},
		},
		Commands: local,
	}
	app.Setup()

	err := app.Run(os.Args)
	if err != nil {
		oo.LogD("%v", err)
		return
	}
}
