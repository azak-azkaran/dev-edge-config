package main

import (
	"bytes"
	"io/ioutil"
	"os"
"strings"
)

type Server struct {
	Id     string
	Name   string
	IP     string
	Status string
}

// Exists reports whether the named file or directory exists.
func FileExists(name string) bool {
	if _, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

func (server Server) SetId(id string) Server {
	server.Id = id
	Sugar.Debug("Setting ID to", server.Id)
	return server
}

func (server Server) SetName(name string) Server {
	server.Name = strings.TrimSpace(name)
	Sugar.Debug("Setting Name to", server.Name)
	return server
}

func (server Server) SetStatus(status string) Server {
	server.Status = status
	Sugar.Debug("Setting Status to", server.Status)
	return server
}

func (server Server) SetIP(ip string) Server {
	sp := strings.Split(ip, ";")
	if len(sp) == 2 {
		sp = strings.Split(sp[1], ",")

	} else if len(sp) > 2 {
		sp = strings.Split(sp[2], ",")
	} else {
		sp = strings.Split(sp[0], ",")
	}

	if len(sp) > 2 {
		server.IP = sp[1]

	} else if len(sp) == 2 {
		server.IP = sp[1]
	} else {
		server.IP = sp[0]
	}

	if strings.Contains(server.IP, "mgmt") {
		server.IP = strings.ReplaceAll(server.IP, "mgmt=", "")
	}
	Sugar.Debug("Setting IP to", server.IP)
	return server
}

func handleServer(line string) *Server {
	var server Server
	parts := strings.Split(line, "|")
	if len(parts) < 4 {
		return nil
	}
	for i, p := range parts {
		if i == 1 {
			server = server.SetId(p)
		}
		if i == 2 {
			server = server.SetName(p)
		}
		if i == 3 {
			server = server.SetStatus(p)
		}
		if i == 4 {
			server = server.SetIP(p)
		}
	}
	return &server
}

func createSSHConfig(servers []*Server, prefix string) bytes.Buffer {
	var buf bytes.Buffer

	buf.WriteString("\nHost "+prefix+".switch*")
	buf.WriteString("\n\tUser awieland")
	buf.WriteString("\n\tForwardAgent yes")
	buf.WriteString("\n\tStrictHostKeyChecking no")
	buf.WriteString("\n\tProxyJump "+prefix+".bastion")
	buf.WriteString("\n\tIdentityFile ~/.ssh/ansible_user.key")
	buf.WriteString("\n")

	buf.WriteString("\nHost "+prefix+".server*")
	buf.WriteString("\n\tUser awieland")
	buf.WriteString("\n\tForwardAgent yes")
	buf.WriteString("\n\tStrictHostKeyChecking no")
	buf.WriteString("\n\tProxyJump "+prefix+".bastion")
	buf.WriteString("\n\tIdentityFile ~/.ssh/ansible_user.key")
	buf.WriteString("\n")

	for _, s := range servers {
		if strings.Contains(s.Name, "server") {
			buf.WriteString("\nHost "+prefix+"." + s.Name)
			buf.WriteString("\n\tHostname" + s.IP)
			buf.WriteString("\n\tProxyCommand ssh "+prefix+".switch0 /sbin/ip vrf exec default busybox nc -w 3000 %h %p")
			buf.WriteString("\n")
		} else if strings.Contains(s.Name, "switch") {
			buf.WriteString("\nHost "+prefix+"." + s.Name)
			buf.WriteString("\n\tHostname" + s.IP)
			buf.WriteString("\n")
		} else if strings.Contains(s.Name, "bastion") {
			buf.WriteString("\nHost "+prefix +"."+ s.Name)
			buf.WriteString("\n\tHostname" + s.IP)
			buf.WriteString("\n\tStrictHostKeyChecking no")
			buf.WriteString("\n\tUser awieland")
			buf.WriteString("\n\tIdentityFile ~/.ssh/openstack_innovo-employee-awieland.key")
			buf.WriteString("\n")
		} else if strings.Contains(s.Name, "customer-workstation") {
			buf.WriteString("\nHost "+prefix +"." + s.Name)
			buf.WriteString("\n\tHostname" + s.IP)
			buf.WriteString("\n\tUser centos")
			buf.WriteString("\n\tForwardAgent yes")
			buf.WriteString("\n\tStrictHostKeyChecking no")
			buf.WriteString("\n\tIdentityFile ~/.ssh/openstack_innovo-employee-awieland.key")
			buf.WriteString("\n")
		} else {
			Sugar.Info("ignoring Server: " + s.Name)
		}
	}
	return buf
}

func writeSSHConfig(path string, content []byte) bool {
	path, err := ReplacePath(path)
	if err != nil {
		return false
	}

		if FileExists(path) {
		Sugar.Warn("removes old ssh file")
		err := os.Rename(path, path+".backup")
		if err != nil {
			Sugar.Error(err)
			return false
		}
	}
	Sugar.Debug("Writing SSH config:\n", string(content))
	err = ioutil.WriteFile(path, content, 0644)
	if err != nil {
		Sugar.Error(err)
		return false
	}
	return true
}

func readServers(content string, prefix string)  ( []byte, []*Server ) {
	split := strings.Split(content, "\n")
	var servers []*Server
	for _,line := range split {
		s:=handleServer(line)
		if s != nil{
			servers = append(servers, s )
		}
	}
	buffer := createSSHConfig(servers, prefix)
	return buffer.Bytes(), servers
}
