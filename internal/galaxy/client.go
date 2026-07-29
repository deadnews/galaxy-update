package galaxy

import (
	"cmp"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"
)

const defaultBaseURL = "https://galaxy.ansible.com/api"

var (
	errNoVersion = errors.New("no versions found")
	errNotFound  = errors.New("not found")
)

// Client queries the Ansible Galaxy API for the latest collection and role versions.
type Client struct {
	http    *http.Client
	baseURL string
}

// NewClient returns a Client pointed at the public Ansible Galaxy API.
func NewClient() *Client {
	return &Client{
		http:    &http.Client{Timeout: 30 * time.Second},
		baseURL: defaultBaseURL,
	}
}

// LatestCollection returns the highest published version of a namespace.name collection.
func (c *Client) LatestCollection(ctx context.Context, name string) (string, error) {
	namespace, collection, ok := strings.Cut(name, ".")
	if !ok {
		return "", fmt.Errorf("invalid collection name %q", name)
	}

	endpoint := c.baseURL + "/v3/plugin/ansible/content/published/collections/index/" +
		url.PathEscape(namespace) + "/" + url.PathEscape(collection) + "/"

	var body struct {
		HighestVersion struct {
			Version string `json:"version"`
		} `json:"highest_version"`
	}
	if err := c.getJSON(ctx, endpoint, &body); err != nil {
		return "", err
	}
	if body.HighestVersion.Version == "" {
		return "", errNoVersion
	}
	return body.HighestVersion.Version, nil
}

// LatestRole returns the highest version of a namespace.name role.
func (c *Client) LatestRole(ctx context.Context, name string) (string, error) {
	namespace, role, ok := strings.Cut(name, ".")
	if !ok {
		return "", fmt.Errorf("invalid role name %q", name)
	}

	query := url.Values{"namespace": {namespace}, "name": {role}}
	endpoint := c.baseURL + "/v1/roles/?" + query.Encode()

	var body struct {
		Results []struct {
			SummaryFields struct {
				Versions []struct {
					Name string `json:"name"`
				} `json:"versions"`
			} `json:"summary_fields"`
		} `json:"results"`
	}
	if err := c.getJSON(ctx, endpoint, &body); err != nil {
		return "", err
	}
	if len(body.Results) == 0 {
		return "", errNotFound
	}

	versions := make([]string, 0, len(body.Results[0].SummaryFields.Versions))
	for _, v := range body.Results[0].SummaryFields.Versions {
		versions = append(versions, v.Name)
	}
	if len(versions) == 0 {
		return "", errNoVersion
	}
	return slices.MaxFunc(versions, compareVersions), nil
}

func (c *Client) getJSON(ctx context.Context, endpoint string, v any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, http.NoBody)
	if err != nil {
		return fmt.Errorf("new request: %w", err)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("get %s: %w", endpoint, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%d %s", resp.StatusCode,
			strings.ToLower(http.StatusText(resp.StatusCode)))
	}
	if err := json.NewDecoder(resp.Body).Decode(v); err != nil {
		return fmt.Errorf("decode %s: %w", endpoint, err)
	}
	return nil
}

// ref identifies one collection or role to look up.
type ref struct {
	kind kind
	name string
}

// lookup is the outcome of looking up one ref.
type lookup struct {
	version string
	err     error
}

// resolve looks up every ref concurrently.
func (c *Client) resolve(ctx context.Context, refs []ref) map[ref]lookup {
	resolved := make(map[ref]lookup, len(refs))

	var mu sync.Mutex
	var wg sync.WaitGroup
	for _, r := range refs {
		wg.Go(func() {
			version, err := c.latest(ctx, r)

			mu.Lock()
			defer mu.Unlock()
			resolved[r] = lookup{version: version, err: err}
		})
	}
	wg.Wait()
	return resolved
}

func (c *Client) latest(ctx context.Context, r ref) (string, error) {
	if r.kind == kindRole {
		return c.LatestRole(ctx, r.name)
	}
	return c.LatestCollection(ctx, r.name)
}

// compareVersions orders versions by dotted-numeric comparison.
func compareVersions(a, b string) int {
	as := strings.Split(strings.TrimPrefix(a, "v"), ".")
	bs := strings.Split(strings.TrimPrefix(b, "v"), ".")
	for i := 0; i < len(as) || i < len(bs); i++ {
		an, bn := segment(as, i), segment(bs, i)
		if an != bn {
			return cmp.Compare(an, bn)
		}
	}
	return 0
}

func segment(parts []string, i int) int {
	if i >= len(parts) {
		return 0
	}
	n, _ := strconv.Atoi(parts[i])
	return n
}
