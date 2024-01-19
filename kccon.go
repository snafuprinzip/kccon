package main

import (
	"flag"
	"log"
	"os"
	"os/user"
	"path"
	"slices"
)

const (
	globalConfigPath   = "/etc/k8s/config"
	localConfigPathExt = ".kube/config"
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

func main() {
	var contexts Contexts
	var clusters Clusters
	var users Users
	var found bool

	// Argumente auslesen
	namespace := flag.String("n", "default", "k8s namespace")
	flag.Parse()

	// Sicherstellen, dass der aufrufende User der k8s Gruppe angehört
	CheckUser()
	homedir, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("Kann Home Verzeichnis nicht ermitteln: %v", err)
	}

	// globale kubeconfig einlesen
	var globalConfig KubeConfig
	globalConfig.Load(globalConfigPath)

	// kein Kontext als Argument angegeben, also Liste ausgeben und ausführung erfolgreich beenden
	if flag.NArg() == 0 {
		// contexte ausgeben
		globalConfig.ListContexts()
		os.Exit(0)
	}

	// Context selektieren
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
		globalConfig.ListContexts()
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
	personalConfig.Save(path.Join(homedir, localConfigPathExt))
}
