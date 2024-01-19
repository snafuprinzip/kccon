package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/user"
	"path"
	"slices"
)

// CheckUser prüft, ob der aktuelle Benutzer der Gruppe k8s angehört und somit berechtigt ist dieses Tool auszuführen.
func CheckUser() {
	allowedGroup, err := user.LookupGroup("k8s")
	if err != nil {
		log.Fatalf("Kann keine k8s Gruppe finden: %v", err)
	}

	currentUser, err := user.Current()
	if err != nil {
		log.Fatalf("Kann den aktuellen Benutzer nicht ermitteln: %v", err)
	}

	ids, err := currentUser.GroupIds()
	if err != nil {
		log.Fatalf("Kann keine Gruppeninformationen zum aktuellen Benutzer finden: %v", err)
	}

	if !slices.Contains(ids, allowedGroup.Gid) {
		log.Fatalf("Benutzer ist nicht berechtigt dieses Programm auszuführen, da er nicht der k8s Gruppe angehört")
	}
}

func ShowUsage() {
	fmt.Fprintf(flag.CommandLine.Output(), "usage: %s [-n namespace] [-g global config] [-p personal config] [context]\n", os.Args[0])
	fmt.Fprintln(flag.CommandLine.Output(), "  no arguments\n\tlists available contexts")
	flag.PrintDefaults()
}

func main() {
	var globalConfig KubeConfig
	var contexts Contexts
	var clusters Clusters
	var users Users
	var found bool

	// Sicherstellen, dass der aufrufende User der k8s Gruppe angehört
	CheckUser()
	homedir, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("Kann Home Verzeichnis nicht ermitteln: %v", err)
	}

	// Argumente auslesen
	namespace := flag.String("n", "default", "k8s namespace")
	localConfigPath := flag.String("p", path.Join(homedir, ".kube", "config"), "local config file path")
	globalConfigPath := flag.String("g", path.Join("/", "etc", "k8s", "config"), "global config file path")
	flag.Usage = ShowUsage
	flag.Parse()

	// globale kubeconfig einlesen
	globalConfig.Load(*globalConfigPath)

	// kein Kontext als Argument angegeben, also Liste ausgeben und ausführung erfolgreich beenden
	if flag.NArg() == 0 {
		// Kontexte ausgeben
		globalConfig.ListContexts(*localConfigPath)
		os.Exit(0)
	}

	// Kontext selektieren
	selectedContext := flag.Arg(0)
	found = false
	for _, con := range globalConfig.Contexts {
		if con.Name == selectedContext {
			contexts = con
			found = true
			break
		}
	}

	if !found {
		globalConfig.ListContexts(*localConfigPath)
		log.Fatalf("Unbekannter Kontext: %s", selectedContext)
	}

	contexts.Context.Namespace = *namespace

	// Cluster zum Kontext selektieren
	found = false
	for _, cl := range globalConfig.Clusters {
		if cl.Name == contexts.Context.Cluster {
			clusters = cl
			found = true
			break
		}
	}

	if !found {
		log.Fatalf("Unbekannter Cluster: %s", contexts.Context.Cluster)
	}

	// User zum Kontext selektieren
	found = false
	for _, us := range globalConfig.Users {
		if us.Name == contexts.Context.User {
			users = us
			found = true
			break
		}
	}

	if !found {
		log.Fatalf("Unbekannter User: %s", contexts.Context.User)
	}

	// Individuelle kubeconfig erstellen
	personalConfig := KubeConfig{
		APIVersion:     "v1",
		Clusters:       []Clusters{clusters},
		Contexts:       []Contexts{contexts},
		CurrentContext: selectedContext,
		Kind:           "Config",
		Preferences:    Preferences{},
		Users:          []Users{users},
	}

	// individuelle kubeconfig schreiben
	personalConfig.Save(*localConfigPath)
}
