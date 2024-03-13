<!--suppress HtmlDeprecatedAttribute -->

<div align="center">
  <h1>KYVE + COSMOVISOR</h1>
</div>

![banner](https://github.com/kyve-org/assets/raw/main/banners/KYSOR.png)

## Why use KYSOR

KYSOR is used to manage and run protocol nodes on the KYVE network. 
It ensures that protocol node runners use the correct versions and are always up to date.

Without KYSOR for every pool the node runner has to build and run the protocol node and runtime containers manually.

**Running nodes with KYSOR has the following benefits:**

- Only use **one** program to run on **every** pool
- Automatically **download** and **run** the correct runtime/protocol version
- Getting the new runtime upgrade during a pool upgrade **automatically** and therefore **don't risk timeout slashes**

### Installation/Update

Get the latest release of the KYSOR binaries [here](https://github.com/KYVENetwork/kyve-rdk/releases?q=kysor&expanded=true)

Set the platform you want to download the binary for:
```bash
PLATFORM="linux-amd64"  # options: linux-amd64, linux-arm64, darwin-amd64, darwin-arm64
```

Download the latest KYSOR binary:
```bash
REPO="https://github.com/KYVENetwork/kyve-rdk"
LATEST_VERSION="$(git -c 'versionsort.suffix=-' \
    ls-remote --exit-code --refs --sort='version:refname' --tags $REPO '*.*.*' \
    | grep kysor \
    | tail --lines=1 \
    | cut --delimiter='/' --fields=3-)"
    
wget "$REPO/releases/download/$LATEST_VERSION"/kysor-"$PLATFORM"
wget "$REPO/releases/download/$LATEST_VERSION"/checksums.txt

sha256sum --check --ignore-missing checksums.txt  # verify the checksum

mv kysor-"$PLATFORM" kysor  # rename the binary (optional move to /usr/local/bin)
chmod +x kysor  # make the binary executable
```

To verify that the KYSOR runs successfully just run

```bash
./kysor version
```

### Initialize KYSOR

Once you have successfully downloaded the KYSOR binary you have to initialize it.

You can use your own RPC/REST endpoints, or use the ones provided by KYVE.

| Network         | Chain ID   | RPC URL                            | REST URL                           |
|-----------------|------------|------------------------------------|------------------------------------|
| Mainnet         | kyve-1     | https://rpc.kyve.network/          | https://api.kyve.network/          |
| Kaon testnet    | kaon-1     | https://rpc.kaon.kyve.network/     | https://api.kaon.kyve.network/     |
| Korellia devnet | korellia-2 | https://rpc.korellia.kyve.network/ | https://api.korellia.kyve.network/ |

```bash
./kysor init    # you can add the --home flag to specify a custom home directory (ex: --home ~/.korellia)
```
This creates a `config.toml` under the following directory: `$HOME/.kysor/` (or the home that you provided via flag). 
You can edit this file manually later on.

**Create your first valaccount**

Now that KYSOR is initialized we move on to the next step. 
For every pool you run on a _valaccount_ has to be created. In our example, we want to run on the Cosmoshub pool. 
A new valaccount with a new mnemonic can be created in the following way:

```bash
./kysor valaccounts create  # then follow the instructions
```

You can also provide the flags directly to the command like this:
```bash
./kysor valaccounts create \
  --name cosmoshub \  # the name of the valaccount
  --storage-priv "$(cat path/to/arweave.json)"  # the private key of the Arweave wallet
```

More help on how to manage valaccounts can be found with `./kysor valaccounts [subcommand] --help`

> 📝 **INFORMATION**<br>
> We recommend to name the valaccount after the pool you want to run on. This makes it easier to manage multiple valaccounts.
> These names are just used locally for config management. <br>
> If you have multiple valaccounts running on the same machine you are required to change the port of the metrics server (if enabled of course) so they don't overlap.


### Run KYSOR

After you have created the required valaccounts you can simply start running the KYSOR with the following command:

```bash
./kysor start # This will guide you through the process of starting the KYSOR
```

If you want to skip the interactive mode you can also start the KYSOR with the following command:

```bash
./kysor start --valaccount cosmoshub --yes
```

To see all available flags you can use the `--help` flag like this:
```bash
./kysor start --help
```

### Run KYSOR with systemd

For the daemon service root-privileges are required during the setup. 
Create a service file. $USER is the Linux user which runs the process. Replace it before you copy the command.

Since the KYSOR can run on multiple pools on one machine we would recommend naming the daemon service after the valaccount name and with a `d` appending to it.
With that you can create multiple service files and control each of them. 
This example shows the service file for our valaccount `cosmoshub`

```bash
tee <<EOF > /dev/null /etc/systemd/system/cosmoshubd.service
[Unit]
Description=KYVE Protocol-Node cosmoshub daemon
After=network-online.target

[Service]
User=$USER
ExecStart=/home/$USER/kysor start --valaccount cosmoshub --yes
Restart=on-failure
RestartSec=3
LimitNOFILE=infinity
EOF
```

Start the daemon

```bash
sudo systemctl enable cosmoshubd
sudo systemctl start cosmoshubd
```

It can be stopped using

```
sudo systemctl stop cosmoshubd
```

You can see its logs with

```
sudo journalctl -u cosmoshubd -f
```

### Stop KYSOR
```bash
./kysor stop
```