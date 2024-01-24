# kccon

kccon is a command line tool for switching kubernetes contexts.

It copies kubernetes contexts from a global kubeconfig file
that contains all kubernetes clusters and can be maintained centrally.

This way the users can share a central kubeconfig without interferring with another users context and namespace while
copying only the really necessary parts into a local user kubeconfig, which can be deleted regulary, e.g. via a cronjob
or by deleting it from your shell logout file.

Access to the global config is limited to members of the _k8s_ group, so the os access rights to the global kubeconfig 
have to be set accordingly.

## Usage

```shell
$ usage:
list contexts:   kccon [-g global config] [-p personal config]
set context:     kccon [-n namespace] [-g global config] [-p personal config] <context>
add new context: [sudo] kccon [-g global config] [-p personal config] -a <file>
remove context:  [sudo] kccon [-g global config] [-p personal config] -r <context>

  -a string
    	add new context from <file> to global config
  -g string
    	global config file <path> (default "/etc/k8s/config")
  -n string
    	k8s <namespace> (default "default")
  -p string
    	local config file <path> (default "/home/mleimenmeier/.kube/config")
  -r string
    	remove <context> from global config
```
If a context is given on the command line the context (and namespace) will be switched and written to the local kubeconfig file.

If no context is given _kccon_ will show a list of the available contexts from the global kubeconfig file, marking the
current context and namespace from the local kubeconfig.

## Installation

build binary and copy it to /usr/local/bin/

```
go build .
sudo cp kccon /usr/local/bin/
sudo chown root:k8s /usr/local/bin/kccon
sudo chmod 0750 /usr/local/bin/kccon
```

create _k8s_ group and add user

```
sudo groupadd k8s
sudo usermod -a -G k8s <username>
```

copy global config and set permissions

```
sudo mkdir /etc/k8s
sudo chown root:k8s /etc/k8s
sudo chmod 0750 /etc/k8s
sudo cp <globalconfig.yaml> /etc/k8s/config
sudo chown root:k8s /etc/k8s/config
sudo chmod 0640 /etc/k8s/config
```
