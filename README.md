# Offline Gcloud Wrapper

This is a simple tool to query Google Cloud to fetch information about VirtualMachines
and store it locally to have offline and quick access to it.


## Why?

Using `gcloud compute instances list` every time you want to check something about a VM
is not very efficient. And add the convenience of fuzzy finding to it


## How to install

You can download the binaries from the [release page](https://github.com/MansoorMajeed/gcq/releases)

### Linux

```
wget https://github.com/MansoorMajeed/gcq/releases/latest/download/gcq-amd64-linux
sudo mv gcq-amd64-linux /usr/local/bin/gcq
```

### MacOS

```
wget https://github.com/MansoorMajeed/gcq/releases/latest/download/gcq-amd64-osx
sudo mv gcq-amd64-osx /usr/local/bin/gcq

```

### Build yourself

```
git clone https://github.com/MansoorMajeed/gcq.git
cd gcq
go build
sudo mv gcq /usr/local/bin/gcq
```


## How to use

Create `~/.gcq.yaml` config file with projects you want to track

```
projects:
    - project1
    - project2
```

Run `gcq update` so that it fetches project data from Google Cloud


### Example Uses

1. Quickly fuzzy find a VM `instance-a-25` in the project `infrastructure-prod`

```
gcq ls infraprod inst25
```

Pass `--ssh` to show the ssh command


gcq searches the name, status, internal and external IP addresses. So you can make use of
any string matching any of those



2. Find which VM has IP address 10.50.1.1

```
gcq ls infrapr 10.50.1.1
```
