#!/usr/bin/env python3
"""
Language Server Protocol (LSP) Specification Parser

This script fetches the LSP 3.17 specification and breaks it down into smaller,
feature-based files for easier navigation and reference.
"""

import argparse
import sys
from pathlib import Path
from urllib.parse import urljoin

import requests
from bs4 import BeautifulSoup


class LSPSpecificationParser:
    def __init__(self, output_dir: str = "lsp_parsed", base_url: str = None):
        self.output_dir = Path(output_dir)
        self.base_url = (
            base_url or "https://microsoft.github.io/language-server-protocol/specifications/lsp/3.17/specification/"
        )
        self.sections = {}
        self.toc = []

    def fetch_specification(self) -> str:
        """Fetch the LSP specification from the web."""
        print(f"Fetching LSP specification from {self.base_url}")
        try:
            response = requests.get(self.base_url, timeout=30)
            response.raise_for_status()
            return response.text
        except requests.RequestException as e:
            print(f"Error fetching specification: {e}")
            sys.exit(1)

    def parse_html_content(self, html_content: str) -> BeautifulSoup:
        """Parse HTML content using BeautifulSoup."""
        return BeautifulSoup(html_content, "html.parser")

    def extract_table_of_contents(self, soup: BeautifulSoup) -> list[dict]:
        """Extract the table of contents from the specification."""
        toc = []

        # Look for the main headings and sub-headings
        headings = soup.find_all(["h1", "h2", "h3", "h4", "h5", "h6"])

        for heading in headings:
            level = int(heading.name[1])  # Extract number from h1, h2, etc.
            text = heading.get_text().strip()
            heading_id = heading.get("id", "")

            # Skip empty headings or those without meaningful content
            if not text or len(text) < 3:
                continue

            toc.append({"level": level, "text": text, "id": heading_id, "element": heading})

        return toc

    def identify_major_sections(self, soup: BeautifulSoup) -> dict[str, dict]:
        """Identify major sections of the specification."""
        sections = {}

        # Define major section patterns
        major_sections = {
            "base_protocol": ["Base Protocol", "baseProtocol"],
            "basic_structures": ["Basic JSON Structures", "basicJsonStructures"],
            "lifecycle": ["Server lifecycle", "Lifecycle Messages", "lifeCycleMessages"],
            "text_synchronization": [
                "Text Document Synchronization",
                "Document Synchronization",
                "textDocument_synchronization",
            ],
            "notebook_synchronization": ["Notebook Document Synchronization", "notebookDocument_synchronization"],
            "language_features": ["Language Features", "languageFeatures"],
            "workspace_features": ["Workspace Features", "workspaceFeatures"],
            "window_features": ["Window Features", "windowFeatures"],
            "miscellaneous": ["Miscellaneous", "Implementation Considerations", "implementationConsiderations"],
            "change_log": ["Change Log", "changeLog"],
            "meta_model": ["Meta Model", "metaModel"],
        }

        # Find all headings
        headings = soup.find_all(["h1", "h2", "h3"])

        for heading in headings:
            text = heading.get_text().strip()
            heading_id = heading.get("id", "")

            # Check if this heading matches any major section
            for section_key, patterns in major_sections.items():
                if any(
                    pattern.lower() in text.lower() or pattern.lower() in heading_id.lower() for pattern in patterns
                ):
                    sections[section_key] = {"title": text, "id": heading_id, "element": heading, "content": []}
                    break

        return sections

    def extract_section_content(self, soup: BeautifulSoup, start_element, end_element=None) -> list:
        """Extract content between two elements."""
        content = []
        current = start_element.next_sibling

        while current and current != end_element:
            if hasattr(current, "name"):  # It's a tag
                content.append(current)
            current = current.next_sibling

        return content

    def extract_language_features(self, soup: BeautifulSoup) -> dict[str, dict]:
        """Extract individual language features from the specification."""
        features = {}

        # Common LSP request/notification patterns
        lsp_patterns = [
            "textDocument/completion",
            "textDocument/hover",
            "textDocument/signatureHelp",
            "textDocument/declaration",
            "textDocument/definition",
            "textDocument/typeDefinition",
            "textDocument/implementation",
            "textDocument/references",
            "textDocument/documentHighlight",
            "textDocument/documentSymbol",
            "textDocument/codeAction",
            "textDocument/codeLens",
            "textDocument/documentLink",
            "textDocument/colorProvider",
            "textDocument/formatting",
            "textDocument/rangeFormatting",
            "textDocument/onTypeFormatting",
            "textDocument/rename",
            "textDocument/prepareRename",
            "textDocument/foldingRange",
            "textDocument/selectionRange",
            "textDocument/publishDiagnostics",
            "textDocument/diagnostic",
            "textDocument/semanticTokens",
            "textDocument/inlayHint",
            "textDocument/inlineValue",
            "textDocument/moniker",
            "textDocument/linkedEditingRange",
            "textDocument/callHierarchy",
            "textDocument/typeHierarchy",
        ]

        # Find headings that contain these patterns
        headings = soup.find_all(["h2", "h3", "h4"])

        for heading in headings:
            text = heading.get_text().strip()
            heading_id = heading.get("id", "")

            # Check if this heading describes a language feature
            for pattern in lsp_patterns:
                if pattern in text.lower() or pattern.replace("/", "_") in heading_id.lower():
                    feature_name = pattern.replace("textDocument/", "").replace("/", "_")
                    features[feature_name] = {"title": text, "id": heading_id, "element": heading, "method": pattern}
                    break

            # Also check for workspace features
            workspace_patterns = [
                "workspace/symbol",
                "workspace/executeCommand",
                "workspace/applyEdit",
                "workspace/didChangeConfiguration",
                "workspace/workspaceFolders",
            ]

            for pattern in workspace_patterns:
                if pattern in text.lower() or pattern.replace("/", "_") in heading_id.lower():
                    feature_name = pattern.replace("workspace/", "workspace_")
                    features[feature_name] = {"title": text, "id": heading_id, "element": heading, "method": pattern}
                    break

        return features

    def create_feature_files(self, soup: BeautifulSoup):
        """Create individual files for each feature."""
        # Create output directory
        self.output_dir.mkdir(parents=True, exist_ok=True)

        # Extract language features
        features = self.extract_language_features(soup)

        # Extract major sections
        sections = self.identify_major_sections(soup)

        # Create files for major sections
        for section_key, section_info in sections.items():
            self.create_section_file(soup, section_key, section_info)

        # Create files for individual language features
        features_dir = self.output_dir / "language_features"
        features_dir.mkdir(exist_ok=True)

        for feature_name, feature_info in features.items():
            self.create_feature_file(soup, feature_name, feature_info, features_dir)

        # Create an index file
        self.create_index_file(sections, features)

    def create_section_file(self, soup: BeautifulSoup, section_key: str, section_info: dict):
        """Create a file for a major section."""
        filename = f"{section_key}.md"
        filepath = self.output_dir / filename

        content = self.extract_section_content_detailed(soup, section_info["element"])

        with open(filepath, "w", encoding="utf-8") as f:
            f.write(f"# {section_info['title']}\n\n")
            f.write(f"**Section ID:** {section_info['id']}\n\n")
            f.write("---\n\n")
            f.write(content)

        print(f"Created: {filepath}")

    def create_feature_file(self, soup: BeautifulSoup, feature_name: str, feature_info: dict, output_dir: Path):
        """Create a file for an individual language feature."""
        filename = f"{feature_name}.md"
        filepath = output_dir / filename

        content = self.extract_section_content_detailed(soup, feature_info["element"])

        with open(filepath, "w", encoding="utf-8") as f:
            f.write(f"# {feature_info['title']}\n\n")
            f.write(f"**Method:** `{feature_info.get('method', 'N/A')}`\n")
            f.write(f"**Section ID:** {feature_info['id']}\n\n")
            f.write("---\n\n")
            f.write(content)

        print(f"Created: {filepath}")

    def extract_section_content_detailed(self, soup: BeautifulSoup, start_element) -> str:
        """Extract detailed content for a section, converting to markdown."""
        content_parts = []
        current = start_element

        # Find the next heading of the same or higher level to know where to stop
        start_level = int(start_element.name[1]) if start_element.name.startswith("h") else 1

        # Get all siblings after the start element
        for sibling in start_element.find_next_siblings():
            if sibling.name and sibling.name.startswith("h"):
                sibling_level = int(sibling.name[1])
                if sibling_level <= start_level:
                    break

            # Convert element to markdown
            content_parts.append(self.element_to_markdown(sibling))

        return "\n".join(content_parts)

    def element_to_markdown(self, element) -> str:
        """Convert an HTML element to markdown format."""
        if not hasattr(element, "name"):
            return str(element).strip()

        if element.name == "p":
            return element.get_text().strip() + "\n"
        elif element.name in ["h1", "h2", "h3", "h4", "h5", "h6"]:
            level = int(element.name[1])
            return "#" * level + " " + element.get_text().strip() + "\n"
        elif element.name == "pre":
            code_content = element.get_text()
            return f"```\n{code_content}\n```\n"
        elif element.name == "code":
            return f"`{element.get_text()}`"
        elif element.name == "ul":
            items = []
            for li in element.find_all("li", recursive=False):
                items.append(f"- {li.get_text().strip()}")
            return "\n".join(items) + "\n"
        elif element.name == "ol":
            items = []
            for i, li in enumerate(element.find_all("li", recursive=False), 1):
                items.append(f"{i}. {li.get_text().strip()}")
            return "\n".join(items) + "\n"
        elif element.name == "table":
            return self.table_to_markdown(element)
        elif element.name == "blockquote":
            lines = element.get_text().strip().split("\n")
            return "\n".join(f"> {line}" for line in lines) + "\n"
        else:
            return element.get_text().strip() + "\n"

    def table_to_markdown(self, table) -> str:
        """Convert an HTML table to markdown format."""
        rows = []

        # Get headers
        headers = []
        header_row = table.find("tr")
        if header_row:
            for th in header_row.find_all(["th", "td"]):
                headers.append(th.get_text().strip())

        if headers:
            rows.append("| " + " | ".join(headers) + " |")
            rows.append("| " + " | ".join(["---"] * len(headers)) + " |")

        # Get data rows
        for tr in table.find_all("tr")[1:]:  # Skip header row
            cells = []
            for td in tr.find_all(["td", "th"]):
                cells.append(td.get_text().strip())
            if cells:
                rows.append("| " + " | ".join(cells) + " |")

        return "\n".join(rows) + "\n"

    def create_index_file(self, sections: dict, features: dict):
        """Create an index file listing all created files."""
        index_path = self.output_dir / "README.md"

        with open(index_path, "w", encoding="utf-8") as f:
            f.write("# LSP Specification - Parsed Files\n\n")
            f.write("This directory contains the Language Server Protocol (LSP) 3.17 specification ")
            f.write("broken down into smaller, feature-based files for easier navigation.\n\n")

            f.write("## Major Sections\n\n")
            for section_key, section_info in sections.items():
                filename = f"{section_key}.md"
                f.write(f"- [{section_info['title']}]({filename})\n")

            f.write("\n## Language Features\n\n")
            for feature_name, feature_info in features.items():
                filename = f"language_features/{feature_name}.md"
                f.write(f"- [{feature_info['title']}]({filename})\n")

            f.write("\n## Original Source\n\n")
            f.write(f"These files were generated from: {self.base_url}\n")
            f.write(f"Generated on: {self.get_current_timestamp()}\n")

        print(f"Created index: {index_path}")

    def get_current_timestamp(self) -> str:
        """Get current timestamp as string."""
        from datetime import datetime

        return datetime.now().strftime("%Y-%m-%d %H:%M:%S")

    def run(self):
        """Main execution method."""
        print("Starting LSP Specification Parser...")

        # Fetch the specification
        html_content = self.fetch_specification()

        # Parse HTML
        soup = self.parse_html_content(html_content)

        # Create feature files
        self.create_feature_files(soup)

        print(f"\nParsing complete! Files created in: {self.output_dir}")
        print(f"Total files created: {len(list(self.output_dir.rglob('*.md')))}")


def main():
    parser = argparse.ArgumentParser(description="Parse LSP specification into feature-based files")
    parser.add_argument("-o", "--output", default="lsp_parsed", help="Output directory (default: lsp_parsed)")
    parser.add_argument(
        "-u",
        "--url",
        default="https://microsoft.github.io/language-server-protocol/specifications/lsp/3.17/specification/",
        help="LSP specification URL",
    )
    parser.add_argument("--local-file", help="Use local HTML file instead of fetching from URL")

    args = parser.parse_args()

    # Create parser instance
    lsp_parser = LSPSpecificationParser(output_dir=args.output, base_url=args.url)

    # If local file is specified, read it instead of fetching
    if args.local_file:
        print(f"Reading local file: {args.local_file}")
        try:
            with open(args.local_file, encoding="utf-8") as f:
                html_content = f.read()
            soup = lsp_parser.parse_html_content(html_content)
            lsp_parser.create_feature_files(soup)
            print(f"\nParsing complete! Files created in: {args.output}")
        except FileNotFoundError:
            print(f"Error: File {args.local_file} not found")
            sys.exit(1)
        except Exception as e:
            print(f"Error reading file: {e}")
            sys.exit(1)
    else:
        # Run normal web-based parsing
        lsp_parser.run()


if __name__ == "__main__":
    main()
