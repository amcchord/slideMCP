# Slide Documentation Tool

The `slide_docs` tool provides access to Slide documentation from docs.slide.tech directly within the MCP agent. This allows LLMs to reference official documentation when answering questions about Slide backup and recovery features.

## Features

### 1. List Documentation Sections
Browse available documentation categories:
```json
{
  "operation": "list_sections"
}
```

Returns main documentation sections like:
- Getting Started
- Slide Console
- Product
- Billing
- API
- Troubleshooting
- Best Practices

### 2. Get Topics in a Section
View specific topics within a documentation section:
```json
{
  "operation": "get_topics",
  "section": "Getting Started"
}
```

### 3. Search Documentation
Search across all documentation for specific keywords:
```json
{
  "operation": "search_docs",
  "query": "backup retention"
}
```

Returns matches from both topic titles and content with previews.

### 4. Get Specific Content
Retrieve detailed documentation for a specific topic:
```json
{
  "operation": "get_content",
  "topic": "API Authentication"
}
```

Returns full markdown-formatted documentation content.

### 5. Get API Reference
Quick access to API reference information:
```json
{
  "operation": "get_api_reference",
  "endpoint": "agents" // optional
}
```

## Usage Examples

### Example 1: Learning about backups
```json
{
  "operation": "search_docs",
  "query": "backup process"
}
```

### Example 2: API Authentication Help
```json
{
  "operation": "get_content",
  "topic": "API Authentication"
}
```

### Example 3: Exploring Network Configuration
```json
{
  "operation": "get_content",
  "topic": "Network Configuration"
}
```

## Integration with Other Tools

The `slide_docs` tool complements other Slide MCP tools by providing:
- Context for API operations
- Best practices for configuration
- Troubleshooting guidance
- Feature explanations

When using other slide_* tools, you can reference documentation to:
- Understand parameter meanings
- Learn about feature capabilities
- Find troubleshooting steps
- Discover best practices

## Implementation Notes

- Documentation content is cached for performance
- Returns markdown-formatted content for easy reading
- Provides both exact content matches and template responses
- Links to docs.slide.tech for the most up-to-date information

## Future Enhancements

Potential improvements could include:
- Real-time documentation fetching from docs.slide.tech
- Version-specific documentation
- Interactive examples
- Video tutorial references
- Integration with support tickets 