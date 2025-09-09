package scraper

import (
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"

	"github.com/gocolly/colly/v2"

	"github.com/paran01d/lddb/internal/models"
)

// LDDBScraper handles scraping LaserDisc information from lddb.com
type LDDBScraper struct {
	collector *colly.Collector
}

// NewLDDBScraper creates a new LDDB scraper
func NewLDDBScraper() *LDDBScraper {
	c := colly.NewCollector()

	// Set user agent to avoid being blocked
	c.UserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36"

	// Add some basic error handling
	c.OnError(func(r *colly.Response, err error) {
		log.Printf("Error scraping %s: %v", r.Request.URL, err)
	})

	return &LDDBScraper{
		collector: c,
	}
}

// LookupByUPC searches lddb.com for LaserDisc information using UPC
func (s *LDDBScraper) LookupByUPC(upc string) (*models.LookupResult, error) {
	result := &models.LookupResult{
		UPC:   upc,
		Found: false,
	}

	// Clean UPC (remove any non-digit characters)
	cleanUPC := regexp.MustCompile(`\D`).ReplaceAllString(upc, "")
	if len(cleanUPC) == 0 {
		result.Error = "Invalid UPC format"
		return result, nil
	}

	// First, search for the UPC to get basic info and the detailed link
	searchURL := fmt.Sprintf("https://www.lddb.com/search.php?UPC=%s", cleanUPC)
	var detailURL string

	// Set up scraping rules for search results
	s.collector.OnHTML("html", func(e *colly.HTMLElement) {
		// Check if we got search results or a direct hit
		pageText := strings.ToLower(e.Text)
		
		// If page contains "no results" or similar, return not found
		if strings.Contains(pageText, "no results") || 
		   strings.Contains(pageText, "not found") ||
		   strings.Contains(pageText, "0 results") {
			result.Found = false
			return
		}

		// Look for LaserDisc information in various page structures
		s.extractLaserDiscInfo(e, result)
		
		// Try to find the detailed link by looking for LaserDisc ID pattern
		lines := strings.Split(e.Text, "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			// Look for pattern like "LV323503WS Ghost and the Darkness, The (1996)"
			ldIdPattern := regexp.MustCompile(`(LV\w+)\s+.+?\((\d{4})\)`)
			if matches := ldIdPattern.FindStringSubmatch(line); len(matches) > 1 {
				ldId := matches[1]
				// Construct detailed URL
				detailURL = fmt.Sprintf("https://www.lddb.com/laserdisc/%s/", ldId)
				break
			}
		}
	})

	// Visit the search URL first
	err := s.collector.Visit(searchURL)
	if err != nil {
		result.Error = fmt.Sprintf("Failed to fetch search results: %v", err)
		return result, nil
	}
	
	s.collector.Wait()

	// If we found a detailed URL and basic info, get more detailed information
	if detailURL != "" && result.Found {
		err := s.getDetailedInfo(detailURL, result)
		if err != nil {
			log.Printf("Warning: Could not fetch detailed info from %s: %v", detailURL, err)
			// Don't fail the whole lookup, just log the warning
		}
	}

	return result, nil
}

// getDetailedInfo fetches detailed information from the LaserDisc's dedicated page
func (s *LDDBScraper) getDetailedInfo(url string, result *models.LookupResult) error {
	// Create a new collector for the detailed page
	detailCollector := colly.NewCollector()
	detailCollector.UserAgent = s.collector.UserAgent
	
	detailCollector.OnHTML("html", func(e *colly.HTMLElement) {
		pageText := e.Text
		
		// Look for director information
		directorPattern := regexp.MustCompile(`(?i)director[:\s]+([^\n\r]+)`)
		if matches := directorPattern.FindStringSubmatch(pageText); len(matches) > 1 {
			director := strings.TrimSpace(matches[1])
			// Clean up common artifacts
			director = strings.TrimSuffix(director, ",")
			director = strings.TrimSuffix(director, ".")
			if len(director) > 0 && len(director) < 100 {
				result.Director = director
			}
		}
		
		// Look for genre information
		genrePattern := regexp.MustCompile(`(?i)genre[:\s]+([^\n\r]+)`)
		if matches := genrePattern.FindStringSubmatch(pageText); len(matches) > 1 {
			genre := strings.TrimSpace(matches[1])
			genre = strings.TrimSuffix(genre, ",")
			genre = strings.TrimSuffix(genre, ".")
			if len(genre) > 0 && len(genre) < 50 {
				result.Genre = genre
			}
		}
		
		// Look for runtime information
		runtimePatterns := []string{
			`(?i)runtime[:\s]+(\d+)\s*min`,
			`(?i)duration[:\s]+(\d+)\s*min`,
			`(?i)running time[:\s]+(\d+)\s*min`,
		}
		
		for _, pattern := range runtimePatterns {
			re := regexp.MustCompile(pattern)
			if matches := re.FindStringSubmatch(pageText); len(matches) > 1 {
				if runtime, err := strconv.Atoi(matches[1]); err == nil {
					result.Runtime = runtime
					break
				}
			}
		}
		
		// Look for cover image
		e.ForEach("img", func(_ int, img *colly.HTMLElement) {
			src := img.Attr("src")
			alt := strings.ToLower(img.Attr("alt"))
			if strings.Contains(src, "cover") || strings.Contains(alt, "cover") || strings.Contains(alt, "laserdisc") {
				if strings.HasPrefix(src, "/") {
					result.CoverImageURL = "https://www.lddb.com" + src
				} else if strings.HasPrefix(src, "http") {
					result.CoverImageURL = src
				}
			}
		})
	})
	
	return detailCollector.Visit(url)
}

// extractLaserDiscInfo extracts LaserDisc information from the HTML
func (s *LDDBScraper) extractLaserDiscInfo(e *colly.HTMLElement, result *models.LookupResult) {
	// Try different selectors to find the information
	// LDDB.com structure might vary, so we'll try multiple approaches

	// Method 1: Look for table-based layout (common in LDDB)
	s.extractFromTable(e, result)

	// Method 2: Look for definition lists
	if !result.Found {
		s.extractFromDefinitionList(e, result)
	}

	// Method 3: Look for specific div structures
	if !result.Found {
		s.extractFromDivStructure(e, result)
	}

	// Method 4: Extract from text patterns (fallback)
	if !result.Found {
		s.extractFromTextPatterns(e, result)
	}
}

// extractFromTable extracts info from table-based layouts
func (s *LDDBScraper) extractFromTable(e *colly.HTMLElement, result *models.LookupResult) {
	// Look for tables containing LaserDisc information
	e.ForEach("table", func(_ int, table *colly.HTMLElement) {
		if result.Found {
			return
		}

		tableText := strings.ToLower(table.Text)
		if strings.Contains(tableText, "title") || strings.Contains(tableText, "laserdisc") {
			// Extract information from table rows
			table.ForEach("tr", func(_ int, row *colly.HTMLElement) {
				s.extractFromTableRow(row, result)
			})

			if result.Title != "" {
				result.Found = true
			}
		}
	})
}

// extractFromTableRow extracts info from a table row
func (s *LDDBScraper) extractFromTableRow(row *colly.HTMLElement, result *models.LookupResult) {
	cells := row.ChildTexts("td")
	if len(cells) < 2 {
		return
	}

	label := strings.ToLower(strings.TrimSpace(cells[0]))
	value := strings.TrimSpace(cells[1])

	switch {
	case strings.Contains(label, "title"):
		result.Title = value
	case strings.Contains(label, "year") || strings.Contains(label, "date"):
		if year, err := s.extractYear(value); err == nil {
			result.Year = year
		}
	case strings.Contains(label, "director") || strings.Contains(label, "directed"):
		result.Director = value
	case strings.Contains(label, "genre") || strings.Contains(label, "category"):
		result.Genre = value
	case strings.Contains(label, "format"):
		result.Format = value
	case strings.Contains(label, "sides"):
		if sides, err := strconv.Atoi(strings.Fields(value)[0]); err == nil {
			result.Sides = sides
		}
	case strings.Contains(label, "runtime") || strings.Contains(label, "duration"):
		if runtime, err := s.extractRuntime(value); err == nil {
			result.Runtime = runtime
		}
	}
}

// extractFromDefinitionList extracts info from definition lists
func (s *LDDBScraper) extractFromDefinitionList(e *colly.HTMLElement, result *models.LookupResult) {
	e.ForEach("dl", func(_ int, dl *colly.HTMLElement) {
		if result.Found {
			return
		}

		terms := dl.ChildTexts("dt")
		definitions := dl.ChildTexts("dd")

		for i, term := range terms {
			if i >= len(definitions) {
				break
			}

			label := strings.ToLower(strings.TrimSpace(term))
			value := strings.TrimSpace(definitions[i])

			s.processLabelValue(label, value, result)
		}

		if result.Title != "" {
			result.Found = true
		}
	})
}

// extractFromDivStructure extracts info from div-based structures
func (s *LDDBScraper) extractFromDivStructure(e *colly.HTMLElement, result *models.LookupResult) {
	// Look for specific div classes or patterns
	e.ForEach("div", func(_ int, div *colly.HTMLElement) {
		if result.Found {
			return
		}

		class := div.Attr("class")
		id := div.Attr("id")
		
		if strings.Contains(strings.ToLower(class), "disc") || 
		   strings.Contains(strings.ToLower(id), "disc") ||
		   strings.Contains(strings.ToLower(class), "title") {
			
			text := div.Text
			s.extractFromText(text, result)
			
			if result.Title != "" {
				result.Found = true
			}
		}
	})
}

// extractFromTextPatterns extracts info using regex patterns on the full page text
func (s *LDDBScraper) extractFromTextPatterns(e *colly.HTMLElement, result *models.LookupResult) {
	text := e.Text
	
	// Look for LDDB-specific format: LV323503WS Ghost and the Darkness, The (1996)LBX/AC3/THX1997-04-15NTSCUSA
	// Note: LDDB uses non-breaking spaces (\u00a0) instead of regular spaces
	lines := strings.Split(text, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		
		// Clean non-breaking spaces that LDDB uses
		cleanLine := strings.ReplaceAll(line, "\u00a0", " ")
		
		// Look for lines that might contain LaserDisc information
		if strings.Contains(strings.ToLower(line), strings.ToLower(result.UPC)) || 
		   (len(result.UPC) > 6 && strings.Contains(line, result.UPC[len(result.UPC)-6:])) ||
		   strings.Contains(line, "LV") {
			
			// Try to extract full information from the LDDB format
			pattern := regexp.MustCompile(`(LV\w+)\s+(.+?)\s*\((\d{4})\)([A-Z/]+)`)
			if matches := pattern.FindStringSubmatch(cleanLine); len(matches) >= 4 {
				result.Title = strings.TrimSpace(matches[2])
				if year, err := strconv.Atoi(matches[3]); err == nil {
					result.Year = year
				}
				result.Format = matches[4]
				result.Found = true
				return
			}
		}
	}
	
	// Generic patterns as fallback
	titlePatterns := []string{
		`Title:\s*(.+?)(?:\n|$)`,
		`TITLE:\s*(.+?)(?:\n|$)`,
	}
	
	for _, pattern := range titlePatterns {
		re := regexp.MustCompile(pattern)
		if matches := re.FindStringSubmatch(text); len(matches) > 1 {
			title := strings.TrimSpace(matches[1])
			if !strings.Contains(strings.ToLower(title), "lddb") && 
			   !strings.Contains(strings.ToLower(title), "search") {
				result.Title = title
				break
			}
		}
	}
	
	// Extract year
	yearPattern := regexp.MustCompile(`(?:Year|Date):\s*(\d{4})`)
	if matches := yearPattern.FindStringSubmatch(text); len(matches) > 1 {
		if year, err := strconv.Atoi(matches[1]); err == nil {
			result.Year = year
		}
	}
	
	if result.Title != "" {
		result.Found = true
	}
}

// extractFromText extracts information from raw text
func (s *LDDBScraper) extractFromText(text string, result *models.LookupResult) {
	lines := strings.Split(text, "\n")
	
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		
		parts := strings.Split(line, ":")
		if len(parts) == 2 {
			label := strings.ToLower(strings.TrimSpace(parts[0]))
			value := strings.TrimSpace(parts[1])
			s.processLabelValue(label, value, result)
		}
	}
}

// processLabelValue processes a label-value pair
func (s *LDDBScraper) processLabelValue(label, value string, result *models.LookupResult) {
	switch {
	case strings.Contains(label, "title"):
		result.Title = value
	case strings.Contains(label, "year") || strings.Contains(label, "date"):
		if year, err := s.extractYear(value); err == nil {
			result.Year = year
		}
	case strings.Contains(label, "director") || strings.Contains(label, "directed"):
		result.Director = value
	case strings.Contains(label, "genre") || strings.Contains(label, "category"):
		result.Genre = value
	case strings.Contains(label, "format"):
		result.Format = value
	case strings.Contains(label, "sides"):
		if sides, err := strconv.Atoi(strings.Fields(value)[0]); err == nil {
			result.Sides = sides
		}
	case strings.Contains(label, "runtime") || strings.Contains(label, "duration"):
		if runtime, err := s.extractRuntime(value); err == nil {
			result.Runtime = runtime
		}
	}
}

// extractYear extracts a year from various text formats
func (s *LDDBScraper) extractYear(text string) (int, error) {
	// Look for 4-digit year
	re := regexp.MustCompile(`(\d{4})`)
	matches := re.FindStringSubmatch(text)
	if len(matches) > 1 {
		return strconv.Atoi(matches[1])
	}
	return 0, fmt.Errorf("no year found")
}

// extractRuntime extracts runtime in minutes from various text formats
func (s *LDDBScraper) extractRuntime(text string) (int, error) {
	// Look for patterns like "120 min", "2:00", "2h 30m"
	text = strings.ToLower(text)
	
	// Pattern: "XXX min"
	if re := regexp.MustCompile(`(\d+)\s*min`); re.MatchString(text) {
		matches := re.FindStringSubmatch(text)
		return strconv.Atoi(matches[1])
	}
	
	// Pattern: "H:MM" or "HH:MM"
	if re := regexp.MustCompile(`(\d+):(\d+)`); re.MatchString(text) {
		matches := re.FindStringSubmatch(text)
		hours, _ := strconv.Atoi(matches[1])
		minutes, _ := strconv.Atoi(matches[2])
		return hours*60 + minutes, nil
	}
	
	// Pattern: "Xh Ym"
	if re := regexp.MustCompile(`(\d+)h\s*(\d+)m`); re.MatchString(text) {
		matches := re.FindStringSubmatch(text)
		hours, _ := strconv.Atoi(matches[1])
		minutes, _ := strconv.Atoi(matches[2])
		return hours*60 + minutes, nil
	}
	
	return 0, fmt.Errorf("no runtime found")
}