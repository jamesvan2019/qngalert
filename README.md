# Listen Qng Node Status

### Alert when node exception 

### Run 

```bash
./qngalert 
```

### rename config.json.example config.json
```json
{
  "email": {
    "host":  "",
    "port": "",
    "user": "",
    "pass": "",
    "to": "",
    "enable":true
  },
  "tg": {
    "token": "",
    "chatID":121321324,
    "enable":true
  },
  "nodes": [
    {
      "rpc": "",
      "user": "",
      "pass": "",
      "gap": 30,
      "alert":{
        "maxAllowErrorTimes": 3,
        "maxBlockTime": 600
      }
    }
  ]
}
```


### Got Telegram ChatId

```bash
# create a telegram bot

add @botfather contact

# token use own
visit https://api.telegram.org/bot<token>/getUpdates
```