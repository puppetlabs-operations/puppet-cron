package main

import (
	"crypto/tls"
	"crypto/x509"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

// Call the installed version of Puppet with a given set of arguments.
// It returns the output to the calling function as a byte array.
func callPuppet(puppetArgs []string) []byte {
	paths := []string{}

	if os.Getenv("PATH") != "" {
		paths = append(paths, os.Getenv("PATH"))
	}

	if runtime.GOOS == "windows" {
		defaultPuppetLocation := string("C:\\Program Files\\Puppet Labs\\Puppet\\bin")
		paths = append(paths, defaultPuppetLocation)
		os.Setenv("PATH", strings.Join(paths, ";"))
	} else {
		defaultPuppetLocation := string("/opt/puppetlabs/bin")
		paths = append(paths, defaultPuppetLocation)
		os.Setenv("PATH", strings.Join(paths, ":"))
	}

	output, err := exec.Command("puppet", puppetArgs...).Output()
	if err != nil {
		log.Fatal(err)
	}

	return output
}

// Use `puppet config print` to retrieve the value of a given key
// from a given section of puppet.conf.
// The value is returned as a string to the callling function.
func puppetConfigGet(section string, key string) string {
	args := []string{"config", "print", "--section", section, key}
	output := callPuppet(args)

	value := string(output[:len(output)-1])
	value = strings.ReplaceAll(value, "\r", "")
	value = strings.ReplaceAll(value, "\n", "")
	return value
}

// use `puppet config set` to set the value of a given key in a given
// section of puppet.conf
func puppetConfigSet(section string, key string, value string) {
	args := []string{"config", "set", "--section", section, key, value}

	callPuppet(args)
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

	log.Printf("Checking if environment '%s' exists on '%s'", environment, server)

	url := fmt.Sprintf("https://%s:%s/puppet/v3/file_metadatas/plugins?environment=%s", server, port, environment)
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

// Check to see if the locally configured puppet environment still exists.
// If it doesn't, revert to the `production` envionment or the one specified
// via `-env`. Once the check and any needed update is complete, run the puppet
// agent.
func main() {
	puppetEnv := flag.String("env", "production", "The Puppet environment to fall back to")
	debugFlag := flag.Bool("debug", false, "Enable debugging information")
	envResetDelay := flag.Int("env-reset-delay", 0, "Delay in minutes before reverting to fall back environment")

	flag.Parse()
	cliArgs := flag.Args()

	log.Print("Starting puppet-runner...")
	environment := puppetConfigGet("agent", "environment")

	puppetArgs := []string{}
	if len(cliArgs) > 0 {
		puppetArgs = cliArgs
	} else {
		puppetArgs = []string{"agent", "--no-daemonize", "--onetime"}
	}

	if environment == "" || !isValidEnvironment(environment) {
		if *envResetDelay != 0 {
			log.Printf("Environment %q is invalid, resetting in %d minutes", environment, *envResetDelay)
			time.Sleep(time.Duration(*envResetDelay) * time.Minute)
		} else {
			log.Printf("Environment %q is invalid; resetting now", environment)
		}
		puppetConfigSet("agent", "environment", *puppetEnv)
	}

	if *debugFlag || os.Getenv("PUPPET_RUNNER_DEBUG") != "" {
		fmt.Printf("Current value of $PATH: %s\n\n", os.Getenv("PATH"))
	}

	log.Printf("Running 'puppet %s'", strings.Join(puppetArgs, " "))

	callPuppet(puppetArgs)
}
