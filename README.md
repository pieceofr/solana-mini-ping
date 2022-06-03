# solana-mini-ping
A mini program which only perform ping from ping api server. config.yaml must in the same folder as executable
    You can modify config.yaml.sample and change its name to config.yaml
## Config.yaml 
+ change SolanaConfig to read your solana cli config

```SolanaConfig: # solana-cli config file
 Dir: /home/marry/
 MainnetPath: config-mainnet-beta.yml
 TestnetPath: config-testnet.yml
 DevnetPath: config-devnet.yml
 ```

 + AlternativeEnpoint
 leave it blank for default 

 ```
 AlternativeEnpoint:
  Mainnet:
  Testnet: http://127.0.0.1 # Change to your own endpoint here
  Devnet:
 ```

 + change receiver
    change to the accout you want to receive transaction
```
PingConfig:
    Receiver: 9qT3WeLV5o3t3GVgCk9A3mpTRjSb9qBvnfrAsVKLhmU5
```

