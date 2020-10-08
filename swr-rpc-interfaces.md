[TOC]

# SWR RPC interfaces

### 1 websocket RPC

Use websocket for communication. After the client connects to the websocket, it will use json format parameters for request and reply. As in the configuration file, the connection address is: 
```
ws://127.0.0.1:3080
```

- json rpc message

  ```
  {
  	"cmd": "info",				 # command 
  	"para": {},					   # para for command
  	"typ": "req",				   # req, resp, push, err
  	"chk": "XXX", 				 # Used to identify different requests
  	"err": {					     # If this object exists, it means an error was returned
          "eno": "xx",		 # error number
          "err": "yy",		 # error message
  	}
  }
  ```


#### 1.1 get chain info

```
req json: {
    "cmd": "info",
    "para": {}
}

rsp json: {
    "cmd": "info",
    "para": {
        "network": "main",			# mainnet
        "version": "1.2.1",			# version of stashware
        "height": 999,				  # latest block height
        "current": "xxx",			  # latest block hash
        "blocks": 999,				  # number of blocks
        "peers": 12,				    # peer count
        "queue_length": 10			# txpool length
    }
}
```

#### 1.2 get chain peers

```
req json: {
    "cmd": "peers",
    "para": {}
}

rsp json: {
    "cmd": "peers",
    "para": {
        "end_point": ["ws://127.0.0.1:1981"],
    }
}
```

#### 1.3 get pending transactions

```
req json: {
    "cmd": "tx_pending",
    "para": {}
}

rsp json: {
    "cmd": "tx_pending",
    "para": {
        "txs": ["xxx"]		
    }
}
```

#### 1.4 get block

```
req json: {
    "cmd": "get_block",				# priority current>block_id>height
    "para": {
        "current": 1,				  # 
        "block_height": 1,		# 
        "block_id": "xxx",		# 
    }
}

rsp json: {
    "cmd": "get_block",
    "para": {
        "indep_hash": "xxx",		
        "hash": "xxx",				  
        "height": 999,				
        "previous_block": "xxx",	
        "nonce": "xxx",				
        "timestamp": 999,			
        "last_retarget": 999,		
        "diff": "xxx",				
        "cumulative_diff": 999,		
        "reward_addr": "xxx",		
        "reward_pool": 999,			
        "weave_size": 999,			
        "block_size": 999,			
        "tags": [{					
            "name": "xxx",
            "value": "xxx",
        }],
        "wallet_list": [{			
            "wallet": "xxx",		
            "quantity": 999,		
            "last_tx": "xxx",		
        }],
        "txs": ["xxx"]				
    }
}
```

#### 1.5 get transaction by id

```
req json: {
    "cmd": "tx_by_id",
    "para": {
        "txid": "xxx"				
    }
}

rsp json: {
    "cmd": "tx_by_id",
    "para": {
        "id": "xxx",				
        "owner": "xxx",				
        "from": "xxx",				
        "last_tx": "xxx",			
        "target": "xxx",			
        "quantity": 999,			
        "reward": 999,				
        "tags": [{					
            "name": "xxx",
            "value": "xxx",
        }],
        "signature": "xxx",			
        "timestamp": 999,			
        "block_indep_hash": "xxx",	
        "confirmations": 10			
        "data_hash": "xxx"			
    }
}
```

#### 1.6 get data

```
req json: {
    "cmd": "tx_data",
    "para": {
        "txid": "xxx"				#
    }
}

rsp json: {
    "cmd": "tx_data",
    "para": "xxx"						# hex encoded
}
```

#### 1.7 get fee

```
req json: {
    "cmd": "price",
    "para": {
    		"bytes": 999,					# lengh of data
        "target": "xxx"				# target address
    }
}

rsp json: {
    "cmd": "price",
    "para": {
    	"amount": 999,					
    }
}
```

#### 1.8 submit block

```
req json: {
    "cmd": "submit_block",
    "para": {
    	"new_block":{
            ...						
    	},
    	"bds": "xxx",				# Block Data Segment
    }
}

rsp json: {
    "cmd": "get_block",
    "para": {
    }
}
```

#### 1.9 submit transaction

```
req json: {
    "cmd": "submit_tx",
    "para": {
        "id": "xxx",				
        "owner": "xxx",				
        "from": "xxx",				
        "last_tx": "xxx",			
        "target": "xxx",			
        "quantity": 999,			
        "reward": 999,				
        "tags": [{					
            "name": "xxx",
            "value": "xxx",
        }],
        "data": "xxx",				
        "data_hash": "xxx"			
        "signature": "xxx",			
    }
}

rsp json: {
    "cmd": "submit_tx",
    "para": {
    }
}

# note
TAGS = (result=name+value+name+value...)

MSGS = data: binary(owner)+binary(target)+binary(sha256(data))+binary(quantity)+binary(reward)+binary(last_tx)+binary(TAGS)

sigdata = RSA(private key，SHA256(MSGS))
signature = hex(sigdata)

id = hex(SHA256(sigdata)) 

ex1:
{
  "id" : "1143b3920638a39d796a588ef80564d322d0dd3544270f93588f8529e295bcba",
  "last_tx" : "",
  "owner" : "30820122300d06092a864886f70d01010105000382010f003082010a0282010100c9215878f3b9801d342dac0702287b27b170e0e005c908e9a5a4e491f3d8349dba57718de7422843525753d049dead294d999ac4a40e1f46c7995d1949bef770f242cd8957a0126dc375209bbfcec421ec45bcdcc6521d3764206500f6527335be27bf898f2452b0a4592f751c827dca94e2dc6385ed4a10a82ed5bbfd3cec395d345697cd487e2a7a26bd3f23a933d75dc3d59ac1170e46f4ef517205e72b16569578f2caeeefbb589fa98f86bfc4ff91dd2ada7e1d562897ced78061245be48914dcc28b2b91eb3806a3b9dfc2d73ef759be7a0ee2d8853e8c035b98613cb965cea2f6775dd34fc049c95b3f2f3757b0934c21b4f861e37f088d4a8aa9a7910203010001",
  "from" : "3Nbzx3SsJubTgiMpzFLYiae2y3BYhprxRz",
  "target" : "3Bh4UCRsnHwt3NLmLymjMY1CYV8SYfJM2P",
  "quantity" : 10000000000,
  "reward" : 2000028814,
  "tags" : [{
      "name" : "name",
      "value" : "value"
    }
  ],
  "data" : "00000068656c6c6f776f726c64",
  "data_hash" : "49c2c5175caf94e7828978eaabd53696708255b572ab1021f2c8e202e1f71e77",
  "signature" : "c741102978b16754fc248524c5d8d45f8691c3bd0c6a67bc7be721abe2108cdcab07f16832c67e9b56cb9d57878e975b0aec2666bd744dc858416bfc258930adc90a774898f564239a9982201dfe1daeadc00c16893814be1de9fe2c89a5190c82314473ddf11635a752dd43cd6a50b0b7fbe7e02f9dfedfc1837c8f74ca6c6260f2bbb01960bc117ad956e5e967c9bc1c73e6144b38d787fb1aef2c561040066b2800cac22b0137108e0c1e55b5e8e5e6e3b298d819bb514d5891b09cc5f4270e657cfccb932dd33f5d1ab5662ed208798910e46b7fcb20091c5052061d28c2e26208cc9d73f979cad1e0c987f37fc9a1ce050fc7b9556571a0de1c899ee069"
}

```

#### 1.10 new wallet

```
req json: {
  "cmd": "wallet_new",
    "para": {
    }
}

rsp json: {
    "cmd": "wallet_new",
    "para": {
    	"address": "xxx",			
    	"public_key": "xxx",		
    	"private_key": "xxx",		
    }
}
```

#### 1.11 get wallet info

```
req json: {
    "cmd": "wallet_by_addr",
    "para": {
    	"address": "xxx",			
    }
}

rsp json: {
    "cmd": "wallet_by_addr",
    "para": {
    	"pending": 999,				
    	"balance": 999,				
    	"last_tx": "xxx",		
    	"public_key": "xxx",	
    }
}
```

#### 1.12 get transactions of address

```
req json: {
    "cmd": "get_addr_txs",
    "para": {
    	"address": "xxx",		
    	"height1": 123,			
    	"height2": 456,			
    }
}

rsp json: {
    "cmd": "get_addr_txs",
    "para": {
    	"data": [
            "xxx"					
    	]
    }
}
```

### 2 REST api

The return of the REST interface is the same as that of the websocket RPC interface, except that the request method is different. The default port is 2080. The following is an example of curl. The braces in the URL indicate variable parameters. For details, please refer to the websocket RPC interface.

#### 2.1 get chain info
```
curl 'http://localhost:2080/info'
```
#### 2.2 get chain peers
```
curl 'http://localhost:2080/peers'
```
#### 2.3 get pending transactions
```
curl 'http://localhost:2080/tx/pending'
```
#### 2.4 get block
```
curl 'http://localhost:2080/block/hash/{block_hash}'	
curl 'http://localhost:2080/block/height/{height}'	
```
#### 2.5 get transaction by id
```
curl 'http://localhost:2080/tx/{txid}'
```
#### 2.6 get data
```
curl 'http://localhost:2080/{txid}.{extension}'
```
#### 2.7 get fee
```
curl 'http://localhost:2080/price/{bytes}/{target}'
```
#### 2.8 submit transaction
```
curl -X POST 'http://localhost:2080/tx' -d '{...}'
```
#### 2.9 new wallet
```
curl 'http://localhost:2080/wallet/new'
```
#### 2.10 get wallet info
```
curl 'http://localhost:2080/wallet/{address}'
```

#### 2.11 get transactions of address
```
curl 'http://localhost:2080/txs/{address}'			
curl 'http://localhost:2080/txs/{address}/123'	
curl 'http://localhost:2080/txs/{address}/123/456'
```
