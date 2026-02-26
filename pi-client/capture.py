#!/usr/bin/env python3
"""
Zetl Pi Client - Capture handwritten notecards and post quotes to Zetl.

Usage:
  python capture.py          # Capture once and exit
  python capture.py --loop   # Wait for Enter between captures

Requirements:
  - fswebcam installed: sudo apt install fswebcam
  - pip install -r requirements.txt
  - .env file with ZETL_URL, API_TOKEN, ANTHROPIC_API_KEY
"""

import argparse
import base64
import json
import os
import subprocess
import sys
import tempfile

import anthropic
import requests
from dotenv import load_dotenv

load_dotenv()

ZETL_URL = os.environ["ZETL_URL"].rstrip("/")
API_TOKEN = os.environ["API_TOKEN"]
ANTHROPIC_API_KEY = os.environ["ANTHROPIC_API_KEY"]
WEBCAM_DEVICE = os.getenv("WEBCAM_DEVICE", "0")

SYSTEM_PROMPT = (
    "You are an OCR assistant. Extract quote fields from a photo of a handwritten notecard. "
    "Return ONLY valid JSON with no extra text, markdown, or code fences. "
    "Schema: {\"quote\": string, \"author\": string, \"book\": string, \"tags\": [string], \"notes\": string}. "
    "Rules: quote, author, and book are required — use \"Unknown\" if genuinely illegible. "
    "tags defaults to [] and notes defaults to \"\" if not present on the card."
)


def capture_image(output_path: str) -> None:
    cmd = [
        "fswebcam",
        "-d", f"/dev/video{WEBCAM_DEVICE}",
        "-r", "1920x1080",
        "--jpeg", "95",
        "-D", "1",
        "--no-banner",
        output_path,
    ]
    result = subprocess.run(cmd, capture_output=True, text=True)
    if result.returncode != 0:
        raise RuntimeError(f"fswebcam failed: {result.stderr.strip()}")


def ocr_image(image_path: str) -> dict:
    with open(image_path, "rb") as f:
        image_data = base64.standard_b64encode(f.read()).decode("utf-8")

    client = anthropic.Anthropic(api_key=ANTHROPIC_API_KEY)
    message = client.messages.create(
        model="claude-haiku-4-5-20251001",
        max_tokens=512,
        system=SYSTEM_PROMPT,
        messages=[
            {
                "role": "user",
                "content": [
                    {
                        "type": "image",
                        "source": {
                            "type": "base64",
                            "media_type": "image/jpeg",
                            "data": image_data,
                        },
                    },
                    {"type": "text", "text": "Extract the quote fields from this notecard."},
                ],
            }
        ],
    )

    raw = message.content[0].text.strip()
    try:
        return json.loads(raw)
    except json.JSONDecodeError as e:
        raise ValueError(f"Claude returned invalid JSON: {e}\nRaw response:\n{raw}") from e


def post_quote(quote_data: dict) -> dict:
    url = f"{ZETL_URL}/api/device/quote"
    headers = {
        "Authorization": f"Bearer {API_TOKEN}",
        "Content-Type": "application/json",
    }
    resp = requests.post(url, json=quote_data, headers=headers, timeout=10)
    resp.raise_for_status()
    return resp.json()


def run_once() -> bool:
    with tempfile.NamedTemporaryFile(suffix=".jpg", delete=False) as tmp:
        image_path = tmp.name

    try:
        print("Capturing image...")
        capture_image(image_path)
        print(f"  Saved to {image_path}")

        print("Running OCR...")
        quote_data = ocr_image(image_path)
        print(f"  Extracted: {json.dumps(quote_data, indent=2)}")

        print("Posting to Zetl...")
        result = post_quote(quote_data)
        print(f"  Success: quote ID {result.get('id', '?')} created")
        return True

    except Exception as e:
        print(f"Error: {e}", file=sys.stderr)
        return False

    finally:
        if os.path.exists(image_path):
            os.unlink(image_path)


def main():
    parser = argparse.ArgumentParser(description="Capture notecard and post quote to Zetl")
    parser.add_argument("--loop", action="store_true", help="Keep running, press Enter to capture")
    args = parser.parse_args()

    if args.loop:
        print("Loop mode — press Enter to capture, Ctrl+C to quit.")
        try:
            while True:
                input("\nPress Enter to capture...")
                run_once()
        except KeyboardInterrupt:
            print("\nExiting.")
    else:
        success = run_once()
        sys.exit(0 if success else 1)


if __name__ == "__main__":
    main()
