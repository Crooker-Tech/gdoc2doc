"""
Setup Google OAuth credentials for gdoc2pdf.

This script handles the OAuth flow for Google Drive/Docs API access.
Run this once to authorize the application and store tokens.

Usage:
    1. Load the OAuth client secret:
       . .\\tools\\load-key.ps1 -Service gdoc2pdf-oauth -Target API
    2. Run: python tools/setup_google_oauth.py
"""
import io
import sys

if sys.stdout.encoding != 'utf-8':
    sys.stdout = io.TextIOWrapper(sys.stdout.buffer, encoding='utf-8')

import os
import subprocess
from pathlib import Path

# Ensure required packages
def ensure_packages():
    packages = ['google-auth-oauthlib', 'google-auth-httplib2', 'google-api-python-client']
    for package in packages:
        try:
            __import__(package.replace('-', '_').split('[')[0])
        except ImportError:
            result = subprocess.run(['uv', 'pip', 'install', package], capture_output=True)
            if result.returncode != 0:
                subprocess.run([sys.executable, '-m', 'pip', 'install', package], check=True)

ensure_packages()

from google_auth_oauthlib.flow import InstalledAppFlow
import json

SCOPES = [
    'https://www.googleapis.com/auth/drive.readonly',
    'https://www.googleapis.com/auth/documents.readonly',
]

# Embedded client configuration (client_secret loaded from credential manager)
CLIENT_CONFIG = {
    "installed": {
        "client_id": "76292221964-1tbb2q1ou4bd6a8doadrsttv3fdrd23a.apps.googleusercontent.com",
        "project_id": "gen-lang-client-0964921467",
        "auth_uri": "https://accounts.google.com/o/oauth2/auth",
        "token_uri": "https://oauth2.googleapis.com/token",
        "auth_provider_x509_cert_url": "https://www.googleapis.com/oauth2/v1/certs",
        "redirect_uris": [
            "http://localhost"
        ]
    }
}


def load_client_secret() -> str:
    """Load OAuth client secret from environment variable."""
    # Environment variable set by: . .\tools\load-key.ps1 -Service gdoc2pdf-oauth -Target API
    secret = os.environ.get('GDOC2PDF-OAUTH_API_KEY')
    if not secret:
        print("Error: GDOC2PDF-OAUTH_API_KEY not set")
        print()
        print("Run this first:")
        print("  . .\\tools\\load-key.ps1 -Service gdoc2pdf-oauth -Target API")
        sys.exit(1)
    return secret


def setup_oauth(output_path: Path) -> dict:
    """Run OAuth flow and save tokens."""
    # Load client secret from credential manager
    client_secret = load_client_secret()

    # Build complete client config with secret
    config = CLIENT_CONFIG.copy()
    config["installed"]["client_secret"] = client_secret

    # Create flow from config dict
    flow = InstalledAppFlow.from_client_config(config, SCOPES)

    print("Opening browser for Google authorization...")
    print("Please sign in and grant access to Google Drive.")
    print()

    credentials = flow.run_local_server(port=0)

    # Build token data to save
    token_data = {
        'token': credentials.token,
        'refresh_token': credentials.refresh_token,
        'token_uri': credentials.token_uri,
        'client_id': credentials.client_id,
        'client_secret': credentials.client_secret,
        'scopes': list(credentials.scopes),
    }

    # Save token to file
    output_path.parent.mkdir(parents=True, exist_ok=True)
    output_path.write_text(json.dumps(token_data, indent=2), encoding='utf-8')

    print(f"Token saved to: {output_path}")
    print()
    print("You can now run gdoc2pdf!")

    return token_data


def main():
    project_root = Path(__file__).parent.parent
    output_path = project_root / 'token.json'

    setup_oauth(output_path)


if __name__ == '__main__':
    main()
