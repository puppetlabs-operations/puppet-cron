package main

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"github.com/danielparks/lockfile"
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
const LockPath = "/var/run/puppet-cron.lock"

func puppetConfigSectionGet(section string, key string) string {
	args := []string{"config", "print"}

	if section != "" {
		args = append(args, "--section", section)
	}

	args = append(args, key)
	output, err := exec.Command(Puppet, args...).Output()
	if err != nil {
		log.Fatalf("Failed: %s %s", Puppet, strings.Join(args, " "))
	}

	return string(output[:len(output)-1])
}

func puppetConfigGet(key string) string {
	return puppetConfigSectionGet("", key)
}

func puppetConfigSectionSet(section string, key string, value string) {
	args := []string{"config", "set", "--section", section, key, value}

	_, err := exec.Command(Puppet, args...).Output()
	if err != nil {
		log.Fatalf("Failed: %s %s", Puppet, strings.Join(args, " "))
	}
}

func httpClient() *http.Client {
	certname := puppetConfigGet("certname")
	ssldir := puppetConfigGet("ssldir")

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

func isValidEnvironment(environment string) bool {
	server := puppetConfigSectionGet("agent", "server")
	port := puppetConfigSectionGet("agent", "masterport")

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
	lockfile.ObtainLock(LockPath)

	environment := puppetConfigSectionGet("agent", "environment")

	if environment == "" || !isValidEnvironment(environment) {
		log.Printf("Environment %q is invalid; resetting", environment)
		puppetConfigSectionSet("agent", "environment", "production")
	}

	err := syscall.Exec(Puppet, []string{"puppet", "agent", "--no-daemonize", "--onetime"}, os.Environ())
	if err != nil {
		log.Fatal(err)
	} else {
		// WTF. exec should only return if it fails
		log.Panic("syscall.Exec returned with no error")
	}
}
