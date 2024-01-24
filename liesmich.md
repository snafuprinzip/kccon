# kccon

kccon ist ein Kommandozeilentool, welches dazu dient, eine globale kubeconfig Datei zentral abzulegen und zu pflegen, 
die alle Kubernetes Cluster enthält, ohne einen gemeinsamen Kontext für alle Benutzer festzulegen.

Hierzu wird bei der Auswahl einen Kontexts mit _kccon_ eine lokale Kopie des gewünschten Kontexts aus der zentralen
Konfiguration in das Homeverzeichnis des Benutzers unter _.kube/config_ abgelegt. 

Vorraussetzung ist, dass der Benutzer der Gruppe _k8s_ angehört und die globale Konfiguration für die _k8s_ Gruppe
lesbar ist.

## Benutzung

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

Wird ein Kontext angegeben, wird der Kontext in die lokale Konfiguration kopiert und ein evtl. ebenfalls angegebener namespace
gesetzt.

Wird kein Kontext angegeben wird eine Liste der verfügbaren Kontexte ausgeben, wobei ein evtl. bereits vorhandener Kontext aus der lokalen Konfiguration markiert
und bei einem breiten Terminal auch der aktuelle Namespace mit ausgegeben wird.

## Installation

Binary erstellen und nach /usr/local/bin kopieren
```
go build .
sudo cp kccon /usr/local/bin/
sudo chown root:k8s /usr/local/bin/kccon
sudo chmod 0750 /usr/local/bin/kccon
```

_k8s_ Gruppe erstellen und Benutzer hinzufügen

```
sudo groupadd k8s
sudo usermod -a -G k8s <username>
```

Globale Konfig zentral ablegen

```
sudo mkdir /etc/k8s
sudo chown root:k8s /etc/k8s
sudo chmod 0750 /etc/k8s
sudo cp <Globalconfig.yaml> /etc/k8s/config
sudo chown root:k8s /etc/k8s/config
sudo chmod 0640 /etc/k8s/config
```
