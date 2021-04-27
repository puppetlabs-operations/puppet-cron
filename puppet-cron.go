package main

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"time"
)

const Puppet = "/opt/puppetlabs/bin/puppet"

// We use `puppet config` to get and set values in
// /etc/puppetlabs/puppet/puppet.conf. We don't touch it directly at all.

func puppetConfigGet(section string, key string) string {
	args := []string{"config", "print", "--section", section, key}

	output, err := exec.Command(Puppet, args...).Output()
	if err != nil {
		log.Fatalf("Failed: %s %s", Puppet, strings.Join(args, " "))
	}

	return string(output[:len(output)-1])
}

func puppetConfigSet(section string, key string, value string) {
	args := []string{"config", "set", "--section", section, key, value}

	_, err := exec.Command(Puppet, args...).Output()
	if err != nil {
		log.Fatalf("Failed: %s %s", Puppet, strings.Join(args, " "))
	}
}

// Create an http.Client that recognizes the Puppet CA, and authenticates with
// node's certificate.
func httpClient() *http.Client {
	certname := puppetConfigGet("agent", "certname")
	ssldir := puppetConfigGet("main", "ssldir")

	clientCertPath := fmt.Sprintf("%s/certs/%s.pem", ssldir, certname)
	clientKeyPath := fmt.Sprintf("%s/private_keys/%s.pem", ssldir, certname)
	caCertPath := fmt.Sprintf("%s/certs/ca.pem", ssldir)

	clientCert, err := tls.LoadX509KeyPair(clientCertPath, clientKeyPath)
	if err != nil {
		log.Fatal(err)
	}

	caCert, err := ioutil.ReadFile(caCertPath)
	if err != nil {
		log.Fatal(err)
	}

	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{clientCert},
		RootCAs:      caCertPool,
	}
	tlsConfig.BuildNameToCertificate()
	transport := &http.Transport{TLSClientConfig: tlsConfig}

	return &http.Client{
		Transport: transport,
		Timeout:   60 * time.Second,
	}
}

// Check that the agent-configured environment still exists on the puppet server.
func isValidEnvironment(environment string) bool {
	server := puppetConfigGet("agent", "server")
	port := puppetConfigGet("agent", "masterport")

	url := fmt.Sprintf("https://%s:%s/puppet/v3/status/test?environment=%s", server, port, environment)
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Panic(err)
	}

	request.Header.Set("Accept", "application/json")

	response, err := httpClient().Do(request)
	if err != nil {
		log.Fatal(err)
	}

	if response.StatusCode == 404 {
		return false
	} else if response.StatusCode < 200 || response.StatusCode >= 300 {
		log.Printf("Unexpected status %q checking environment %q. Keeping environment.", response.Status, environment)
	}

	return true
}

func main() {
	environment := puppetConfigGet("agent", "environment")

	if environment == "" || !isValidEnvironment(environment) {
		log.Printf("Environment %q is invalid; resetting", environment)
		puppetConfigSet("agent", "environment", "production")
	}

	err := syscall.Exec(Puppet, []string{"puppet", "agent", "--no-daemonize", "--onetime"}, os.Environ())
	if err != nil {
		log.Fatal(err)
	} else {
		// WTF. exec should only return if it fails
		log.Panic("syscall.Exec returned with no error")
	}
}
