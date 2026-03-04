package services

import (
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	refCacheDuration = int64(3_600_000) // 1 hour in ms
	docsBaseURL      = "https://raw.githubusercontent.com/citizenfx/fivem-docs/master/content/docs/game-references/"
	imgBaseURL       = "https://docs.fivem.net"
)

// ────────────────────────────────────────────────
// Data types
// ────────────────────────────────────────────────

// Blip represents a map blip icon
type Blip struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	ImageURL string `json:"imageUrl"`
}

// BlipColor represents a blip tint/color
type BlipColor struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Hex  string `json:"hex"`
}

// BlipsResponse wraps blips and blip colors
type BlipsResponse struct {
	Blips  []Blip      `json:"blips"`
	Colors []BlipColor `json:"colors"`
}

// Checkpoint represents a checkpoint type
type Checkpoint struct {
	ID       string `json:"id"`
	Label    string `json:"label"`
	Section  string `json:"section"` // "standard" | "type-44-46"
	ImageURL string `json:"imageUrl"`
}

// DataFile represents a data file type entry
type DataFile struct {
	Key         string `json:"key"`
	FileType    string `json:"fileType"`
	RootElement string `json:"rootElement"`
	Mounter     string `json:"mounter"`
	Example     string `json:"example"`
}

// GameEvent represents a client-side game event
type GameEvent struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// GamerTagComponent represents a gamer tag (head display) component
type GamerTagComponent struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// HUDColor represents a HUD color entry
type HUDColor struct {
	Index int    `json:"index"`
	Name  string `json:"name"`
	R     int    `json:"r"`
	G     int    `json:"g"`
	B     int    `json:"b"`
	A     int    `json:"a"`
	Hex   string `json:"hex"`
}

// Marker represents a 3D world marker type
type Marker struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	ImageURL string `json:"imageUrl"`
}

// NetGameEvent represents a net game event ID
type NetGameEvent struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// PedModel represents a pedestrian model
type PedModel struct {
	Name       string `json:"name"`
	Category   string `json:"category"`
	Props      int    `json:"props"`
	Components int    `json:"components"`
	ImageURL   string `json:"imageUrl"`
}

// PickupHash represents a pickup hash entry
type PickupHash struct {
	Name string `json:"name"`
	Hash string `json:"hash"`
}

// WeaponModel represents a weapon model entry
type WeaponModel struct {
	Name         string   `json:"name"`
	Hash         string   `json:"hash"`
	ModelHashKey string   `json:"modelHashKey"`
	DLC          string   `json:"dlc"`
	Description  string   `json:"description"`
	ImageURL     string   `json:"imageUrl"`
	Group        string   `json:"group"`
	Components   []string `json:"components"`
	Tints        []string `json:"tints"`
}

// Zone represents a map zone
type Zone struct {
	ID          int    `json:"id"`
	ZoneNameID  string `json:"zoneNameId"`
	ZoneName    string `json:"zoneName"`
	Description string `json:"description"`
}

// ────────────────────────────────────────────────
// Service
// ────────────────────────────────────────────────

// GameReferencesService fetches and caches game reference data from the FiveM docs
type GameReferencesService struct {
	client *http.Client
	cache  map[string]*refCacheEntry
	mu     sync.RWMutex
}

type refCacheEntry struct {
	data     interface{}
	cachedAt int64
}

// NewGameReferencesService creates a new GameReferencesService instance
func NewGameReferencesService() *GameReferencesService {
	return &GameReferencesService{
		cache:  make(map[string]*refCacheEntry),
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

func (s *GameReferencesService) getCache(key string) interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if e, ok := s.cache[key]; ok {
		if time.Now().UnixMilli()-e.cachedAt < refCacheDuration {
			return e.data
		}
	}
	return nil
}

func (s *GameReferencesService) setCache(key string, data interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.cache[key] = &refCacheEntry{data: data, cachedAt: time.Now().UnixMilli()}
}

func (s *GameReferencesService) fetchText(path string) (string, error) {
	req, err := http.NewRequest("GET", docsBaseURL+path, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "FixFX-Core/1.0")
	resp, err := s.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to fetch %s: %w", path, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP %d fetching %s", resp.StatusCode, path)
	}
	b, err := io.ReadAll(resp.Body)
	return string(b), err
}

// ────────────────────────────────────────────────
// Blips
// ────────────────────────────────────────────────

// GetBlips returns all blip data
func (s *GameReferencesService) GetBlips() ([]Blip, []BlipColor, error) {
	const key = "blips"
	if cached := s.getCache(key); cached != nil {
		r := cached.(*BlipsResponse)
		return r.Blips, r.Colors, nil
	}
	content, err := s.fetchText("blips.md")
	if err != nil {
		return nil, nil, err
	}
	blips, colors := parseBlips(content)
	s.setCache(key, &BlipsResponse{Blips: blips, Colors: colors})
	return blips, colors, nil
}

func parseBlips(content string) ([]Blip, []BlipColor) {
	divRe := regexp.MustCompile(`(?s)<div class="blip">.*?</span></div>`)
	srcRe := regexp.MustCompile(`src="/blips/([^"]+)"`)
	idRe := regexp.MustCompile(`<strong>(\d+)</strong>`)
	altRe := regexp.MustCompile(`alt="([^"]+)"`)

	var blips []Blip
	seen := make(map[int]bool)

	for _, div := range divRe.FindAllString(content, -1) {
		idMatch := idRe.FindStringSubmatch(div)
		altMatch := altRe.FindStringSubmatch(div)
		srcMatch := srcRe.FindStringSubmatch(div)
		if idMatch == nil || altMatch == nil || srcMatch == nil {
			continue
		}
		id, _ := strconv.Atoi(idMatch[1])
		if seen[id] {
			continue
		}
		seen[id] = true
		blips = append(blips, Blip{
			ID:       id,
			Name:     altMatch[1],
			ImageURL: imgBaseURL + "/blips/" + srcMatch[1],
		})
	}

	// Parse blip colors: look for a table section with | ID | Name | Hex |
	colors := parseBlipColors(content)
	return blips, colors
}

func parseBlipColors(content string) []BlipColor {
	// Find color table section after "## Blip colors" or similar heading
	colorSection := content
	if idx := strings.Index(content, "Blip color"); idx >= 0 {
		colorSection = content[idx:]
	}

	var colors []BlipColor
	rowRe := regexp.MustCompile(`\|\s*(\d+)\s*\|\s*([^|]+?)\s*\|\s*(#[0-9a-fA-F]+)\s*\|`)
	for _, m := range rowRe.FindAllStringSubmatch(colorSection, -1) {
		id, _ := strconv.Atoi(m[1])
		colors = append(colors, BlipColor{
			ID:   id,
			Name: strings.TrimSpace(m[2]),
			Hex:  strings.TrimSpace(m[3]),
		})
	}
	return colors
}

// ────────────────────────────────────────────────
// Checkpoints
// ────────────────────────────────────────────────

// GetCheckpoints returns all checkpoint type data
func (s *GameReferencesService) GetCheckpoints() ([]Checkpoint, error) {
	const key = "checkpoints"
	if cached := s.getCache(key); cached != nil {
		return cached.([]Checkpoint), nil
	}
	content, err := s.fetchText("checkpoints.md")
	if err != nil {
		return nil, err
	}
	checkpoints := parseCheckpoints(content)
	s.setCache(key, checkpoints)
	return checkpoints, nil
}

func parseCheckpoints(content string) []Checkpoint {
	divRe := regexp.MustCompile(`(?s)<div class="checkpoint">.*?</span></div>`)
	srcRe := regexp.MustCompile(`src="/checkpoints/([^"]+)"`)
	idRe := regexp.MustCompile(`<strong>([^<]+)</strong>`)
	labelRe := regexp.MustCompile(`(?s)<strong>[^<]+</strong>(.*?)</span>`)

	// Detect section boundary — "Type 44" appears after standard section
	splitIdx := strings.Index(content, "Checkpoint Type 44")
	standardContent := content
	variant44Content := ""
	if splitIdx >= 0 {
		standardContent = content[:splitIdx]
		variant44Content = content[splitIdx:]
	}

	var result []Checkpoint

	parse := func(src, section string) {
		for _, div := range divRe.FindAllString(src, -1) {
			srcMatch := srcRe.FindStringSubmatch(div)
			idMatch := idRe.FindStringSubmatch(div)
			if srcMatch == nil || idMatch == nil {
				continue
			}
			label := ""
			if lm := labelRe.FindStringSubmatch(div); lm != nil {
				label = strings.TrimSpace(stripHTML(lm[1]))
			}
			result = append(result, Checkpoint{
				ID:       idMatch[1],
				Label:    label,
				Section:  section,
				ImageURL: imgBaseURL + "/checkpoints/" + srcMatch[1],
			})
		}
	}

	parse(standardContent, "standard")
	parse(variant44Content, "type-44-46")
	return result
}

// ────────────────────────────────────────────────
// Data Files
// ────────────────────────────────────────────────

// GetDataFiles returns all data file type entries
func (s *GameReferencesService) GetDataFiles() ([]DataFile, error) {
	const key = "data-files"
	if cached := s.getCache(key); cached != nil {
		return cached.([]DataFile), nil
	}
	content, err := s.fetchText("data-files.md")
	if err != nil {
		return nil, err
	}
	files := parseDataFiles(content)
	s.setCache(key, files)
	return files, nil
}

func parseDataFiles(content string) []DataFile {
	// Each row starts with | and has a span with the key
	rowRe := regexp.MustCompile(`(?m)^\|[^|]*<span[^>]*>` + "`" + `([^` + "`" + `]+)` + "`" + `</span>[^|]*\|([^|]*)\|([^|]*)\|([^|]*)\|([^|]*)`)
	keyRe := regexp.MustCompile("`([^`]+)`")

	var files []DataFile
	for _, m := range rowRe.FindAllStringSubmatch(content, -1) {
		key := m[1]
		if key == "" {
			if km := keyRe.FindStringSubmatch(m[0]); km != nil {
				key = km[1]
			}
		}
		if key == "" {
			continue
		}
		files = append(files, DataFile{
			Key:         strings.TrimSpace(key),
			FileType:    strings.TrimSpace(stripHTML(m[2])),
			RootElement: strings.TrimSpace(stripHTML(m[3])),
			Mounter:     strings.TrimSpace(stripHTML(m[4])),
			Example:     strings.TrimSpace(stripHTML(m[5])),
		})
	}

	// Fallback: just extract all span ids
	if len(files) == 0 {
		spanRe := regexp.MustCompile(`<span\s+id="([^"]+)">[^<]*` + "`([^`]+)`")
		for _, m := range spanRe.FindAllStringSubmatch(content, -1) {
			files = append(files, DataFile{Key: m[2]})
		}
	}
	return files
}

// ────────────────────────────────────────────────
// Game Events
// ────────────────────────────────────────────────

// GetGameEvents returns all game event entries
func (s *GameReferencesService) GetGameEvents() ([]GameEvent, error) {
	const key = "game-events"
	if cached := s.getCache(key); cached != nil {
		return cached.([]GameEvent), nil
	}
	content, err := s.fetchText("game-events.md")
	if err != nil {
		return nil, err
	}
	events := parseMarkdownTable2Col(content, func(col1, col2 string) (string, string) {
		return col1, col2
	})
	result := make([]GameEvent, 0, len(events))
	for _, e := range events {
		result = append(result, GameEvent{Name: e[0], Description: e[1]})
	}
	s.setCache(key, result)
	return result, nil
}

// ────────────────────────────────────────────────
// Gamer Tags
// ────────────────────────────────────────────────

// GetGamerTagComponents returns all gamer tag component entries
func (s *GameReferencesService) GetGamerTagComponents() ([]GamerTagComponent, error) {
	const key = "gamer-tags"
	if cached := s.getCache(key); cached != nil {
		return cached.([]GamerTagComponent), nil
	}
	content, err := s.fetchText("gamer-tags.md")
	if err != nil {
		return nil, err
	}
	rows := parseMarkdownTable2Col(content, func(col1, col2 string) (string, string) {
		return col1, col2
	})
	result := make([]GamerTagComponent, 0, len(rows))
	for _, r := range rows {
		id, _ := strconv.Atoi(strings.TrimSpace(r[0]))
		result = append(result, GamerTagComponent{
			ID:   id,
			Name: strings.TrimSpace(r[1]),
		})
	}
	s.setCache(key, result)
	return result, nil
}

// ────────────────────────────────────────────────
// HUD Colors
// ────────────────────────────────────────────────

// GetHUDColors returns all HUD color entries
func (s *GameReferencesService) GetHUDColors() ([]HUDColor, error) {
	const key = "hud-colors"
	if cached := s.getCache(key); cached != nil {
		return cached.([]HUDColor), nil
	}
	content, err := s.fetchText("hud-colors.md")
	if err != nil {
		return nil, err
	}
	colors := parseHUDColors(content)
	s.setCache(key, colors)
	return colors, nil
}

func parseHUDColors(content string) []HUDColor {
	// Match table rows: <tr><td class="color" style="--color: rgba(R, G, B, A)" id="hud_colour_…">…</td><td>INDEX</td><td>NAME</td><td>…rgba…</td></tr>
	trRe := regexp.MustCompile(`(?s)<tr>.*?</tr>`)
	rgbaStyleRe := regexp.MustCompile(`--color:\s*rgba\(\s*(\d+),\s*(\d+),\s*(\d+),\s*([\d.]+)\)`)
	indexRe := regexp.MustCompile(`(?s)</td>\s*<td[^>]*>\s*(\d+)\s*</td>`)
	nameRe := regexp.MustCompile(`(?s)</td>\s*<td[^>]*>\s*(\d+)\s*</td>\s*<td[^>]*>\s*([A-Z_0-9]+)\s*</td>`)
	hexRe := regexp.MustCompile(`<abbr title="(#[0-9a-fA-F]+)"`)

	var colors []HUDColor
	for _, row := range trRe.FindAllString(content, -1) {
		if !strings.Contains(row, `class="color"`) {
			continue
		}
		rgbaM := rgbaStyleRe.FindStringSubmatch(row)
		indexM := indexRe.FindStringSubmatch(row)
		nameM := nameRe.FindStringSubmatch(row)
		if rgbaM == nil || indexM == nil || nameM == nil {
			continue
		}
		r, _ := strconv.Atoi(rgbaM[1])
		g, _ := strconv.Atoi(rgbaM[2])
		b, _ := strconv.Atoi(rgbaM[3])
		aFloat, _ := strconv.ParseFloat(rgbaM[4], 64)
		a := int(aFloat * 255)
		if aFloat <= 1.0 && aFloat != 0 {
			a = int(aFloat * 255)
		} else if aFloat > 1.0 {
			a = int(aFloat)
		}
		idx, _ := strconv.Atoi(indexM[1])
		name := nameM[2]
		hex := ""
		if hm := hexRe.FindStringSubmatch(row); hm != nil {
			hex = hm[1]
		}
		colors = append(colors, HUDColor{
			Index: idx,
			Name:  name,
			R:     r,
			G:     g,
			B:     b,
			A:     a,
			Hex:   hex,
		})
	}
	return colors
}

// ────────────────────────────────────────────────
// Markers
// ────────────────────────────────────────────────

// GetMarkers returns all marker type data
func (s *GameReferencesService) GetMarkers() ([]Marker, error) {
	const key = "markers"
	if cached := s.getCache(key); cached != nil {
		return cached.([]Marker), nil
	}
	content, err := s.fetchText("markers.md")
	if err != nil {
		return nil, err
	}
	markers := parseMarkers(content)
	s.setCache(key, markers)
	return markers, nil
}

func parseMarkers(content string) []Marker {
	divRe := regexp.MustCompile(`(?s)<div class="marker">.*?</span></div>`)
	srcRe := regexp.MustCompile(`src="/markers/([^"]+)"`)
	idRe := regexp.MustCompile(`<strong>(\d+)</strong>`)
	// Name is the text node after the last <br> before </span>
	nameRe := regexp.MustCompile(`(?s)<strong>\d+</strong><br>([^<]+)</span>`)

	var markers []Marker
	for _, div := range divRe.FindAllString(content, -1) {
		srcM := srcRe.FindStringSubmatch(div)
		idM := idRe.FindStringSubmatch(div)
		if srcM == nil || idM == nil {
			continue
		}
		id, _ := strconv.Atoi(idM[1])
		name := ""
		if nm := nameRe.FindStringSubmatch(div); nm != nil {
			name = strings.TrimSpace(nm[1])
		}
		markers = append(markers, Marker{
			ID:       id,
			Name:     name,
			ImageURL: imgBaseURL + "/markers/" + srcM[1],
		})
	}
	return markers
}

// ────────────────────────────────────────────────
// Net Game Events
// ────────────────────────────────────────────────

// GetNetGameEvents returns all net game event entries
func (s *GameReferencesService) GetNetGameEvents() ([]NetGameEvent, error) {
	const key = "net-game-events"
	if cached := s.getCache(key); cached != nil {
		return cached.([]NetGameEvent), nil
	}
	content, err := s.fetchText("net-game-events.md")
	if err != nil {
		return nil, err
	}
	events := parseCEnum(content)
	result := make([]NetGameEvent, 0, len(events))
	for i, name := range events {
		result = append(result, NetGameEvent{ID: i, Name: name})
	}
	s.setCache(key, result)
	return result, nil
}

// ────────────────────────────────────────────────
// Ped Models
// ────────────────────────────────────────────────

// GetPedModels returns all ped model entries
func (s *GameReferencesService) GetPedModels() ([]PedModel, error) {
	const key = "ped-models"
	if cached := s.getCache(key); cached != nil {
		return cached.([]PedModel), nil
	}
	content, err := s.fetchText("ped-models.md")
	if err != nil {
		return nil, err
	}
	peds := parsePedModels(content)
	s.setCache(key, peds)
	return peds, nil
}

func parsePedModels(content string) []PedModel {
	// Section headings in the ped models page use <h2> or ## headings
	// Track current category as we scan divs
	divRe := regexp.MustCompile(`(?s)<div class="model">.*?</span></div>`)
	srcRe := regexp.MustCompile(`src="/peds/([^"]+)"`)
	nameRe := regexp.MustCompile(`<strong>([^<]+)</strong>`)
	propsRe := regexp.MustCompile(`(\d+)\s+prop`)
	compRe := regexp.MustCompile(`(\d+)\s+component`)

	// Category headings: lines like "## Ambient female" or <h2>...</h2>
	catRe := regexp.MustCompile(`(?m)^##\s+(.+)$|<h2>([^<]+)</h2>`)

	type boundedDiv struct {
		start int
		end   int
		text  string
	}

	// Build list of (position, category) for section headings
	type catPos struct {
		pos  int
		name string
	}
	var catPositions []catPos
	for _, m := range catRe.FindAllStringSubmatchIndex(content, -1) {
		pos := m[0]
		name := ""
		if m[2] >= 0 {
			name = strings.TrimSpace(content[m[2]:m[3]])
		} else if m[4] >= 0 {
			name = strings.TrimSpace(content[m[4]:m[5]])
		}
		if name != "" {
			catPositions = append(catPositions, catPos{pos: pos, name: name})
		}
	}

	getCategoryAt := func(pos int) string {
		cat := ""
		for _, cp := range catPositions {
			if cp.pos <= pos {
				cat = cp.name
			} else {
				break
			}
		}
		return cat
	}

	var peds []PedModel
	for _, m := range divRe.FindAllStringIndex(content, -1) {
		div := content[m[0]:m[1]]
		srcM := srcRe.FindStringSubmatch(div)
		nameM := nameRe.FindStringSubmatch(div)
		if srcM == nil || nameM == nil {
			continue
		}
		props := 0
		if pm := propsRe.FindStringSubmatch(div); pm != nil {
			props, _ = strconv.Atoi(pm[1])
		}
		comps := 0
		if cm := compRe.FindStringSubmatch(div); cm != nil {
			comps, _ = strconv.Atoi(cm[1])
		}
		category := getCategoryAt(m[0])
		peds = append(peds, PedModel{
			Name:       nameM[1],
			Category:   category,
			Props:      props,
			Components: comps,
			ImageURL:   imgBaseURL + "/peds/" + srcM[1],
		})
	}
	return peds
}

// ────────────────────────────────────────────────
// Pickup Hashes
// ────────────────────────────────────────────────

// GetPickupHashes returns all pickup hash entries
func (s *GameReferencesService) GetPickupHashes() ([]PickupHash, error) {
	const key = "pickup-hashes"
	if cached := s.getCache(key); cached != nil {
		return cached.([]PickupHash), nil
	}
	content, err := s.fetchText("pickup-hashes.md")
	if err != nil {
		return nil, err
	}
	pickups := parsePickupHashes(content)
	s.setCache(key, pickups)
	return pickups, nil
}

func parsePickupHashes(content string) []PickupHash {
	// C enum entries:  PICKUP_NAME = 1234567,
	re := regexp.MustCompile(`(?m)^\s*(PICKUP_[A-Z0-9_]+)\s*=\s*(\d+)`)
	var hashes []PickupHash
	for _, m := range re.FindAllStringSubmatch(content, -1) {
		hashes = append(hashes, PickupHash{Name: m[1], Hash: m[2]})
	}
	return hashes
}

// ────────────────────────────────────────────────
// Weapon Models
// ────────────────────────────────────────────────

// GetWeaponModels returns all weapon model entries
func (s *GameReferencesService) GetWeaponModels() ([]WeaponModel, error) {
	const key = "weapon-models"
	if cached := s.getCache(key); cached != nil {
		return cached.([]WeaponModel), nil
	}
	content, err := s.fetchText("weapon-models.md")
	if err != nil {
		return nil, err
	}
	weapons := parseWeaponModels(content)
	s.setCache(key, weapons)
	return weapons, nil
}

func parseWeaponModels(content string) []WeaponModel {
	// Group headings: <h2>Heavy group</h2> or <h2>Pistol group</h2>
	groupRe := regexp.MustCompile(`(?i)<h2>([^<]+?)\s*group</h2>|<h2>([^<]+?)</h2>`)
	// Individual weapon divs
	weaponBlockRe := regexp.MustCompile(`(?s)<div class="weapon">(.*?)(?:<div class="weapon">|</div>\s*</div>\s*</div>)`)

	type groupPos struct {
		pos  int
		name string
	}
	var groups []groupPos
	for _, m := range groupRe.FindAllStringSubmatchIndex(content, -1) {
		name := ""
		if m[2] >= 0 {
			name = strings.TrimSpace(content[m[2]:m[3]])
		} else if m[4] >= 0 {
			name = strings.TrimSpace(content[m[4]:m[5]])
		}
		if name != "" {
			groups = append(groups, groupPos{pos: m[0], name: name})
		}
	}

	getGroupAt := func(pos int) string {
		gname := ""
		for _, g := range groups {
			if g.pos <= pos {
				gname = g.name
			} else {
				break
			}
		}
		return gname
	}

	nameRe := regexp.MustCompile(`<strong>Name:</strong>\s*([^<]+)`)
	hashRe := regexp.MustCompile(`<strong>Hash:</strong>\s*([^\s<]+)`)
	modelHashRe := regexp.MustCompile(`<strong>Model Hash Key:</strong>\s*([^\s<]+)`)
	dlcRe := regexp.MustCompile(`<strong>DLC:</strong>\s*([^\s<]+)`)
	descRe := regexp.MustCompile(`(?s)<strong>Description:</strong>\s*(.*?)</span>`)
	imgRe := regexp.MustCompile(`<img\s+src="/weapons/([^"]+)"`)
	liRe := regexp.MustCompile(`<li[^>]*>([^<]+)</li>`)

	var weapons []WeaponModel
	for _, m := range weaponBlockRe.FindAllStringSubmatchIndex(content, -1) {
		block := content[m[2]:m[3]]
		pos := m[0]

		nameM := nameRe.FindStringSubmatch(block)
		if nameM == nil {
			continue
		}

		hashM := hashRe.FindStringSubmatch(block)
		modelHashM := modelHashRe.FindStringSubmatch(block)
		dlcM := dlcRe.FindStringSubmatch(block)
		imgM := imgRe.FindStringSubmatch(block)

		w := WeaponModel{
			Name:  strings.TrimSpace(nameM[1]),
			Group: getGroupAt(pos),
		}
		if hashM != nil {
			w.Hash = strings.TrimSpace(hashM[1])
		}
		if modelHashM != nil {
			w.ModelHashKey = strings.TrimSpace(modelHashM[1])
		}
		if dlcM != nil {
			w.DLC = strings.TrimSpace(dlcM[1])
		}
		if imgM != nil {
			w.ImageURL = imgBaseURL + "/weapons/" + imgM[1]
		}
		if descM := descRe.FindStringSubmatch(block); descM != nil {
			w.Description = strings.TrimSpace(stripHTML(descM[1]))
		}

		// Components: find the components div
		compIdx := strings.Index(block, `class="components"`)
		tintsIdx := strings.Index(block, `class="tints"`)
		if compIdx >= 0 {
			end := len(block)
			if tintsIdx > compIdx {
				end = tintsIdx
			}
			compSection := block[compIdx:end]
			for _, li := range liRe.FindAllStringSubmatch(compSection, -1) {
				w.Components = append(w.Components, strings.TrimSpace(li[1]))
			}
		}

		// Tints
		if tintsIdx >= 0 {
			tintSection := block[tintsIdx:]
			for _, li := range liRe.FindAllStringSubmatch(tintSection, -1) {
				w.Tints = append(w.Tints, strings.TrimSpace(li[1]))
			}
		}

		weapons = append(weapons, w)
	}
	return weapons
}

// ────────────────────────────────────────────────
// Zones
// ────────────────────────────────────────────────

// GetZones returns all zone entries
func (s *GameReferencesService) GetZones() ([]Zone, error) {
	const key = "zones"
	if cached := s.getCache(key); cached != nil {
		return cached.([]Zone), nil
	}
	content, err := s.fetchText("zones.md")
	if err != nil {
		return nil, err
	}
	zones := parseZones(content)
	s.setCache(key, zones)
	return zones, nil
}

func parseZones(content string) []Zone {
	var zones []Zone
	lines := strings.Split(content, "\n")
	inTable := false
	headerPassed := false

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "|") {
			if inTable {
				inTable = false
				headerPassed = false
			}
			continue
		}
		// Separator row
		if strings.Contains(line, "---") {
			inTable = true
			headerPassed = true
			continue
		}
		if !headerPassed {
			inTable = true
			continue
		}
		cols := splitTableRow(line)
		if len(cols) < 4 {
			continue
		}
		id, err := strconv.Atoi(strings.TrimSpace(cols[0]))
		if err != nil {
			continue
		}
		zones = append(zones, Zone{
			ID:          id,
			ZoneNameID:  strings.TrimSpace(cols[1]),
			ZoneName:    strings.TrimSpace(cols[2]),
			Description: strings.TrimSpace(cols[3]),
		})
	}
	return zones
}

// ────────────────────────────────────────────────
// Shared helpers
// ────────────────────────────────────────────────

// parseCEnum extracts ordered names from a C enum block
func parseCEnum(content string) []string {
	// Find enum block between { and }
	start := strings.Index(content, "{")
	end := strings.LastIndex(content, "}")
	if start < 0 || end <= start {
		return nil
	}
	body := content[start+1 : end]

	var names []string
	nameRe := regexp.MustCompile(`(?m)^\s*([A-Z][A-Z0-9_]+)\s*(?:,|$)`)
	for _, m := range nameRe.FindAllStringSubmatch(body, -1) {
		names = append(names, strings.TrimSpace(m[1]))
	}
	return names
}

// parseMarkdownTable2Col parses a markdown table with at least 2 columns, skipping the header.
// Returns [][2]string where [0] is col1, [1] is col2.
func parseMarkdownTable2Col(content string, transform func(string, string) (string, string)) [][2]string {
	var rows [][2]string
	lines := strings.Split(content, "\n")
	headerPassed := false

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "|") {
			continue
		}
		if strings.Contains(line, "---") {
			headerPassed = true
			continue
		}
		if !headerPassed {
			continue
		}
		cols := splitTableRow(line)
		if len(cols) < 2 {
			continue
		}
		c1 := strings.TrimSpace(cols[0])
		c2 := strings.TrimSpace(cols[1])
		if c1 == "" {
			continue
		}
		if transform != nil {
			c1, c2 = transform(c1, c2)
		}
		rows = append(rows, [2]string{c1, c2})
	}
	return rows
}

// splitTableRow splits a markdown table row by | and trims empty leading/trailing cells
func splitTableRow(line string) []string {
	line = strings.TrimSpace(line)
	line = strings.Trim(line, "|")
	parts := strings.Split(line, "|")
	result := make([]string, len(parts))
	for i, p := range parts {
		result[i] = strings.TrimSpace(p)
	}
	return result
}

// stripHTML removes all HTML tags from a string
func stripHTML(s string) string {
	tagRe := regexp.MustCompile(`<[^>]+>`)
	result := tagRe.ReplaceAllString(s, "")
	// Unescape common HTML entities
	result = strings.ReplaceAll(result, "&amp;", "&")
	result = strings.ReplaceAll(result, "&lt;", "<")
	result = strings.ReplaceAll(result, "&gt;", ">")
	result = strings.ReplaceAll(result, "&quot;", `"`)
	result = strings.ReplaceAll(result, "&#39;", "'")
	result = strings.ReplaceAll(result, "&nbsp;", " ")
	return strings.TrimSpace(result)
}

// ────────────────────────────────────────────────
// Generic pagination helper
// ────────────────────────────────────────────────

// RefQuery holds common query parameters for game reference endpoints
type RefQuery struct {
	Search string
	Limit  int
	Offset int
}

// RefMetadata describes pagination context
type RefMetadata struct {
	Total   int    `json:"total"`
	Limit   int    `json:"limit"`
	Offset  int    `json:"offset"`
	HasMore bool   `json:"hasMore"`
	Search  string `json:"search,omitempty"`
}
