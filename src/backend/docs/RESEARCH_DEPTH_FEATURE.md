# Research Pipeline Advanced Configuration

The Research Pipeline template now supports different depth levels and optional Perplexica search settings that significantly change the execution behavior and resource allocation.

## Depth Levels

### Shallow (shallow)
**Best for:** Quick overviews, time-sensitive research, basic fact-checking
- **Search Results:** Up to 3 results
- **Analysis Strategy:** Quick reasoning with 3 iterations
- **Summary Style:** Concise (max 300 words)
- **Timeout:** 3 minutes
- **Agent Names:** "Quick Web Search", "Quick Analysis", "Brief Summary"

### Medium (medium) - Default
**Best for:** Balanced research, most general use cases
- **Search Results:** Up to 5 results  
- **Analysis Strategy:** Balanced reasoning with 5 iterations
- **Summary Style:** Executive summary (max 600 words)
- **Timeout:** 5 minutes
- **Agent Names:** "Web Search", "Analysis", "Summary"

### Deep (deep)
**Best for:** Comprehensive research, academic work, detailed analysis
- **Search Results:** Up to 10 results
- **Analysis Strategy:** Comprehensive reasoning with 8 iterations
- **Summary Style:** Detailed analysis (max 1000 words)
- **Timeout:** 10 minutes
- **Agent Names:** "Comprehensive Web Search", "Deep Analysis", "Detailed Summary"

## Implementation Details

### Dynamic Configuration
The depth parameter dynamically configures:

1. **WebSearchAgent**
   - `max_results` parameter controls search breadth
   - Agent name and description reflect the search scope

2. **ReasoningAgent**
   - `strategy` parameter: "quick", "balanced", or "comprehensive"
   - `max_iterations` parameter controls analysis depth
   - Agent name and description reflect the reasoning approach

3. **SummarizerAgent**
   - `style` parameter: "concise", "executive", or "detailed"
   - `max_length` parameter controls summary length
   - Agent name and description reflect the summary style

4. **Chain Configuration**
   - `timeout` parameter adjusted based on expected execution time
   - Chain name includes depth level for clarity
   - Chain description explains the depth setting

### Usage Example

```json
{
  "topic": "artificial intelligence trends 2024",
  "depth": "deep"
}
```

This will create a research chain named "Research (Deep): artificial intelligence trends 2024" that:
- Searches for up to 10 web results
- Performs comprehensive analysis with 8 reasoning iterations  
- Creates a detailed summary up to 1000 words
- Has a 10-minute timeout to accommodate the deeper analysis

## Optional Perplexica Settings

### Focus Mode (Optional)
When provided, controls how Perplexica performs the search:
- **webSearch** (default): Standard web search
- **academicSearch**: Scholarly and academic sources
- **newsSearch**: Recent news articles  
- **youtubeSearch**: Video content from YouTube
- **redditSearch**: Community discussions from Reddit
- **wolfram**: Computational and factual queries

### Optimization Mode (Optional)  
When provided, controls Perplexica's search optimization:
- **speed**: Faster results with basic analysis
- **balanced** (default): Good balance of speed and quality
- **quality**: More thorough analysis, slower results

## Usage Examples

### Basic Research (defaults only)
```json
{
  "topic": "artificial intelligence trends 2024",
  "depth": "medium"
}
```

### Academic Research with Deep Analysis
```json
{
  "topic": "machine learning in healthcare",
  "depth": "deep",
  "focus_mode": "academicSearch",
  "optimization_mode": "quality"
}
```

### Quick News Research
```json
{
  "topic": "latest AI developments",
  "depth": "shallow", 
  "focus_mode": "newsSearch",
  "optimization_mode": "speed"
}
```

## Benefits

- **Resource Optimization:** Shallow research uses fewer resources and completes faster
- **Quality Scaling:** Deep research provides more thorough analysis for complex topics
- **Search Customization:** Optional Perplexica settings allow targeting specific source types
- **User Control:** Users can choose the appropriate depth and search focus for their needs
- **Clear Expectations:** Agent names and timeouts communicate what to expect
- **Flexible Defaults:** Medium depth with balanced optimization provides good defaults
- **Backward Compatibility:** Existing configurations continue to work without changes