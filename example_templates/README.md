# Template Examples Folder

This folder contains various template examples for use with the `rss2email` tool. Each template is designed to format and display RSS feed items in a specific way. Below, we provide instructions on how to download a YouTube template and a summary of the placeholders used in the template.

## YouTube Template

### Downloading the Template

To use the YouTube template, follow these steps to download it into a local file named `email.tmpl`:

```bash
curl -o email.tmpl https://raw.githubusercontent.com/skx/rss2email/master/example_templates/youtube.txt
```

### Template Summary

The YouTube template is designed to handle YouTube RSS feeds that use the Atom format. The description field in YouTube feeds is not parsed by `gofeed`, and it is stored in the extensions. The template discovers the field by looking for the description name.

#### Placeholders:

1. `{{.RSSItem.Author.Name}}`: Represents the author name.
2. `{{.RSSItem.Published}}`: Represents the published date.
3. `{{.RSSItem.Extensions}}`: Represents the rest of the fields that were not parsed. They are saved using the Extensions type[^1].

The description has been retrieved by digging into the `{{.RSSItem.Extensions}}` attribute.

For more details about the Extensions type and default mappings in `gofeed`, refer to the following documentation:
- [Extensions Type Documentation](https://pkg.go.dev/github.com/mmcdole/gofeed@v1.2.1/extensions#Extensions)[^1]
- [gofeed Default Mappings Documentation](https://pkg.go.dev/github.com/mmcdole/gofeed@v1.2.1#readme-default-mappings)[^2]

[^1]: [Extensions Type Documentation](https://pkg.go.dev/github.com/mmcdole/gofeed@v1.2.1/extensions#Extensions)
[^2]: [gofeed Default Mappings Documentation](https://pkg.go.dev/github.com/mmcdole/gofeed@v1.2.1#readme-default-mappings)

### Feed Format

These feeds can be obtained with the following URL format:

```plaintext
https://www.youtube.com/feeds/videos.xml?channel_id=${CHANNEL_ID}
```

Replace `CHANNEL_ID` with the actual channel ID, which can be retrieved from the channel URL (for instance: `https://www.youtube.com/channel/${CHANNEL_ID}`).
