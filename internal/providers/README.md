# AI Provider Manager

The AI Provider Manager enables automatic failover between multiple AI providers with intelligent error handling and regional restrictions management.

## Features

- **Automatic Failover**: Seamlessly switches between providers when one becomes unavailable
- **Regional Restrictions Handling**: Automatically routes around providers blocked in certain regions (e.g., Gemini in Russia)
- **Circuit Breaker Pattern**: Prevents repeated requests to consistently failing providers
- **Priority-Based Selection**: Uses configurable priorities to determine provider order
- **Error Classification**: Distinguishes between different types of errors for appropriate handling

## Error Handling

The provider manager handles several types of errors:

- **403 Forbidden (Access Denied)**: Typically indicates regional restrictions; immediately switches to next provider
- **500 Internal Server Error**: Indicates provider infrastructure failure; switches to next provider
- **429 Too Many Requests**: Rate limiting; treats as critical error and switches providers
- **Connection/Timeout Errors**: Network issues; treated as critical errors

## Configuration

The provider manager supports multiple providers with configurable priorities:

```go
providerConfig := &providers.MultiProviderConfig{
    Providers: []providers.ProviderConfig{
        {
            Name:     "gemini",
            APIKey:   cfg.LLM.APIKey,
            Regions:  []string{"!RU"}, // Not available in Russia
            Priority: 1,               // Highest priority when available
            Enabled:  true,
        },
        {
            Name:     "openai-mock",
            Regions:  []string{"RU", "US", "EU"}, // Available globally
            Priority: 2,                          // Lower priority
            Enabled:  true,
        },
        {
            Name:     "yandex-gpt-mock",
            Regions:  []string{"RU", "CIS"}, // Available in Russia/CIS
            Priority: 3,                     // Lowest priority
            Enabled:  true,
        },
    },
}
```

## Usage

The provider manager implements the same `LLMClient` interface as individual providers, so it can be used as a drop-in replacement:

```go
// Initialize the provider manager
providerManager := providers.NewProviderManager(providerConfig)

// Register providers
providerManager.RegisterProvider("gemini", geminiClient)
providerManager.RegisterProvider("openai", openaiClient)
providerManager.RegisterProvider("yandex", yandexClient)

// Use as normal LLMClient
response, err := providerManager.Summarize(ctx, parts)
```

## Regional Handling for Russia

For Russian deployments, the system automatically detects when Gemini is blocked (returns 403 errors) and switches to OpenAI or Yandex GPT providers. The configuration prioritizes locally available providers for optimal performance and compliance.

## Fallback Behavior

When all providers fail, the system returns a "no AI models currently working" error, allowing the application to handle the situation gracefully (e.g., by queuing requests for later processing or informing the user).