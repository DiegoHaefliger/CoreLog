package ingest

import (
	"bufio"
	"log"
	"os/exec"
	"strings"
	"sync"
	"time"
)

type DockerCollector struct {
	parser    *Parser
	following map[string]bool
	mu        sync.Mutex
}

func StartDockerCollector(parser *Parser) {
	dc := &DockerCollector{
		parser:    parser,
		following: make(map[string]bool),
	}
	go dc.run()
}

func (dc *DockerCollector) run() {
	for {
		dc.discoverContainers()
		time.Sleep(10 * time.Second)
	}
}

func (dc *DockerCollector) discoverContainers() {
	cmd := exec.Command("docker", "ps", "--format", "{{.Names}}")
	out, err := cmd.Output()
	if err != nil {
		log.Println("CoreLog Docker: erro ao listar containers:", err)
		return
	}

	for _, name := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}

		dc.mu.Lock()
		already := dc.following[name]
		if !already {
			dc.following[name] = true
		}
		dc.mu.Unlock()

		if !already {
			go dc.tailContainer(name)
		}
	}
}

func (dc *DockerCollector) tailContainer(name string) {
	defer func() {
		dc.mu.Lock()
		delete(dc.following, name)
		dc.mu.Unlock()
		log.Printf("CoreLog Docker: container %s encerrado ou removido", name)
	}()

	log.Printf("CoreLog Docker: iniciando coleta do container %s", name)

	cmd := exec.Command("docker", "logs", "--follow", "--tail", "0", name)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Printf("CoreLog Docker: erro no pipe stdout de %s: %v", name, err)
		return
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		log.Printf("CoreLog Docker: erro no pipe stderr de %s: %v", name, err)
		return
	}

	if err := cmd.Start(); err != nil {
		log.Printf("CoreLog Docker: erro ao iniciar coleta de %s: %v", name, err)
		return
	}

	// Muitos apps Node/Java escrevem no stderr — captura os dois
	go func() {
		scanner := bufio.NewScanner(stderr)
		scanner.Buffer(make([]byte, 1024*1024), 1024*1024)
		for scanner.Scan() {
			dc.parser.ParseLineFromSource(scanner.Text(), name)
		}
	}()

	scanner := bufio.NewScanner(stdout)
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024)
	for scanner.Scan() {
		dc.parser.ParseLineFromSource(scanner.Text(), name)
	}

	cmd.Wait()
}
