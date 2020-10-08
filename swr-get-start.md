[TOC]

- Download and unzip the stashware package
- Use ./swr to run the main stashware program
- Open the browser and use http://127.0.0.1:1080 to access the web wallet

- Specification of configuration file swr.conf

```
[global]
data_dir = "./data"           # the directory where all data is stored
txs_dirs = ["./data/txs/"]
network=1

[chain]
relay_txcost_per_byte = 1000
check_points = []             
genesis="genesis.dat"         # genesis file name

[gateway]
listen = ":2080"              # gateway port, default 2080

[node]
listen = ":3080"              # sync port, default 3080

init_peers = ["ws://192.168.1.30:3080", "ws://192.168.1.40:13080"]
reject_ips = ["192.168.1.60"]    # black list

[miner]
core = 8                      # how many core to mine
wallet="swr_wallet.dat"       # miner's wallet file name

[web_wallet]
listen = "127.0.0.1:1080"     # access point of web wallet


```

-  Instructions for using the command line tool swrcli
./swrcli -h
```
NAME:
   swr client - A swr client application

USAGE:
   swrcli [global options] command [command options] [arguments...]

COMMANDS:
   info             current blockchain info
   peers            current connecting peers address
   tx_pending       tx id in memory pool
   get_block        get block by hash or height, default return the latest block
   tx_by_id         get transaction by txid
   tx_data          get tx data, which show in base58 encoded
   price            get cost of saving data to blockchain
   wallet_new       new wallet and return private-key, public-key, address, which are encoded by base58
   wallet_by_addr   get address wallet info
   submit_block     submit new block to swr blockchain
   submit_tx        send new tx to swr blockchain
   get_wallet_list  get the wallet list in block
   get_addr_txs     get address txs id in [height1, height2]
   help, h          Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --host value  (default: "127.0.0.1")
   --port value  (default: 3080)
   --help, -h    show help (default: false)
```
