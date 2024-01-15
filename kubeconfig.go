package main

import (
	"fmt"
	"golang.org/x/term"
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
