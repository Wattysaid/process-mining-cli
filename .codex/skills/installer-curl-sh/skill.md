# installer-curl-sh

```yaml
{
  "skill": {
    "name": "installer-curl-sh",
    "version": "S1.00",
    "date": "2026-01-11",
    "purpose": "Create a safe, user-friendly curl installer that installs the correct binary, verifies checksums, and updates PATH.",
    "when_to_use": "Use when implementing scripts/install.sh, release artefacts, checksums, or self-update.",
    "inputs": [
      "INSTALLATION_AND_RELEASE.md"
    ],
    "outputs": [
      "install.sh + docs + release workflow"
    ],
    "checklist": [
      "Detect OS/arch correctly and map to release artefacts",
      "Check for required tools (curl, unzip/tar, ca-certificates)",
      "Verify SHA256 checksums before installing",
      "Install to ~/.local/bin and update shell rc safely",
      "Print clear next steps including OPENAI_API_KEY setup"
    ],
    "references": [
      "GitHub Releases best practice",
      "Shell script security guidance"
    ]
  }
}
```

## Guidance
Create a safe, user-friendly curl installer that installs the correct binary, verifies checksums, and updates PATH.

## Checklist
- Detect OS/arch correctly and map to release artefacts
- Check for required tools (curl, unzip/tar, ca-certificates)
- Verify SHA256 checksums before installing
- Install to ~/.local/bin and update shell rc safely
- Print clear next steps including OPENAI_API_KEY setup
