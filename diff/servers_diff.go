package diff

import (
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/tufin/oasdiff/utils"
)

// ServersDiff describes the changes between a pair of sets of encoding objects: https://swagger.io/specification/#server-object
type ServersDiff struct {
	Added    utils.StringList `json:"added,omitempty" yaml:"added,omitempty"`
	Deleted  utils.StringList `json:"deleted,omitempty" yaml:"deleted,omitempty"`
	Modified ModifiedServers  `json:"modified,omitempty" yaml:"modified,omitempty"`
}

// ModifiedServers is map of server names to their respective diffs
type ModifiedServers map[string]*ServerDiff

// Empty indicates whether a change was found in this element
func (diff *ServersDiff) Empty() bool {
	if diff == nil {
		return true
	}

	return len(diff.Added) == 0 &&
		len(diff.Deleted) == 0 &&
		len(diff.Modified) == 0
}

func newServersDiff() *ServersDiff {
	return &ServersDiff{
		Added:    utils.StringList{},
		Deleted:  utils.StringList{},
		Modified: ModifiedServers{},
	}
}

func getServersDiff(config *Config, pServers1, pServers2 *openapi3.Servers) *ServersDiff {
	diff := getServersDiffInternal(config, pServers1, pServers2)

	if diff.Empty() {
		return nil
	}

	return diff
}

func getServersDiffInternal(config *Config, pServers1, pServers2 *openapi3.Servers) *ServersDiff {

	result := newServersDiff()

	servers1 := derefServers(pServers1)
	servers2 := derefServers(pServers2)

	for _, server1 := range servers1 {
		if server2 := findServer(server1, servers2); server2 != nil {
			diff, err := getServerDiff(config, server1, server2)
			if err != nil {
				return nil
			}
			if !diff.Empty() {
				result.Modified[server1.URL] = diff
			}
		} else {
			result.Deleted = append(result.Deleted, server1.URL)
		}
	}

	for _, server2 := range servers2 {
		if server1 := findServer(server2, servers1); server1 == nil {
			result.Added = append(result.Added, server2.URL)
		}
	}

	return result
}

func derefServers(servers *openapi3.Servers) openapi3.Servers {
	if servers == nil {
		return openapi3.Servers{}
	}

	return *servers
}

func findServer(server1 *openapi3.Server, servers2 openapi3.Servers) *openapi3.Server {
	// TODO: optimize with a map
	for _, server2 := range servers2 {
		if server2.URL == server1.URL {
			return server2
		}
	}

	return nil
}

func (diff *ServersDiff) getSummary() *SummaryDetails {
	return &SummaryDetails{
		Added:    len(diff.Added),
		Deleted:  len(diff.Deleted),
		Modified: len(diff.Modified),
	}
}
