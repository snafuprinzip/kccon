package main

import (
	"flag"
	"fmt"
	"golang.org/x/term"
	"gopkg.in/yaml.v3"
	"log"
	"os"
	"os/user"
	"path"
	"slices"
	"sort"
)

type KubeConfig struct {
	APIVersion     string      `yaml:"apiVersion"`
	Clusters       []Clusters  `yaml:"clusters"`
	Contexts       []Contexts  `yaml:"contexts"`
	CurrentContext string      `yaml:"current-context"`
	Kind           string      `yaml:"kind"`
	Preferences    Preferences `yaml:"preferences"`
	Users          []Users     `yaml:"users"`
}
type Cluster struct {
	CertificateAuthorityData string `yaml:"certificate-authority-data"`
	Server                   string `yaml:"server"`
}
type Clusters struct {
	Cluster Cluster `yaml:"cluster"`
	Name    string  `yaml:"name"`
}
type Context struct {
	Cluster   string `yaml:"cluster"`
	User      string `yaml:"user"`
	Namespace string `yaml:"namespace,omitempty"`
}
type Contexts struct {
	Context Context `yaml:"context"`
	Name    string  `yaml:"name"`
}
type Preferences struct {
}
type User struct {
	ClientCertificateData string `yaml:"client-certificate-data"`
	ClientKeyData         string `yaml:"client-key-data"`
}
type Users struct {
	Name string `yaml:"name"`
	User User   `yaml:"user"`
}

// ListContexts listet alle Kontexte der globalen kubeconfig unter /etc/k8s/config auf
func (c *KubeConfig) ListContexts() {
	activeSign := " "
	sortedContexts := c.Contexts
	sort.Slice(sortedContexts, func(i, j int) bool {
		return sortedContexts[i].Name < sortedContexts[j].Name
	})

	var wide bool
	if term.IsTerminal(0) {
		w, _, _ := term.GetSize(0)
		if w > 175 {
			wide = true
		}
	}

	if wide {
		fmt.Printf("%-3s %-35s %-50s %-60s %-30s\n", "CUR", "NAME", "CLUSTER", "AUTHINFO", "NAMESPACE")
	} else {
		fmt.Printf("%-3s %-35s %-50s\n", "CUR", "NAME", "CLUSTER")
	}

	for _, con := range sortedContexts {
		{
			if con.Name == c.CurrentContext {
				activeSign = " * "
			} else {
				activeSign = "   "
			}
			if wide {
				fmt.Printf("%-3s %-35s %-50s %-60s %-30s\n", activeSign, con.Name, con.Context.Cluster, con.Context.User, con.Context.Namespace)
			} else {
				fmt.Printf("%-3s %-35s %-50s\n", activeSign, con.Name, con.Context.Cluster)
			}
		}
	}
}

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
		log.Fatalf("Kann keine Gruppeinformationen zum aktuellen Benutzer finden: %v", err)
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

	// globale kubeconfig einlesen
	var globalConfig KubeConfig
	srcFile, err := os.ReadFile("/etc/k8s/config")
	if err != nil {
		log.Printf("srcFile.Get err   #%v ", err)
	}

	err = yaml.Unmarshal(srcFile, &globalConfig)
	if err != nil {
		log.Fatalf("Unmarshal: %v", err)
	}

	// kein Kontext als Argument angegeben, also Liste ausgeben
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
		log.Fatalf("Unbekannter Kontext: %s\n", selectedContext)
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
		log.Fatalf("Unbekannter Cluster: %s\n", contexts.Context.Cluster)
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
		log.Fatalf("Unbekannter User: %s\n", contexts.Context.User)
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
	out, err := yaml.Marshal(&personalConfig)
	if err != nil {
		log.Fatalf("Marshal: %v", err)
	}

	homedir, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("Cannot determine home directory: %v", err)
	}

	err = os.WriteFile(path.Join(homedir, ".kube", "config"), out, 0600)
	if err != nil {
		log.Fatalf("Kann individuelle kubeconnfig nicht schreiben: %v\n", err)
	}
}
