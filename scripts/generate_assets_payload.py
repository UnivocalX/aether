#!/usr/bin/env python3
"""
Generate test JSON payloads for AssetBatchPostPayload testing.
Usage: python generate_assets.py [num_assets] [extra_size]
"""

import json
import random
import string
import sys


def generate_hex_checksum():
    """Generate a random 64-character hexadecimal string."""
    return "".join(random.choices("0123456789abcdef", k=64))


def generate_display():
    """Generate a random display name up to 120 characters."""
    length = random.randint(10, 120)
    return "".join(
        random.choices(string.ascii_letters + string.digits + " ", k=length)
    ).strip()


def generate_extra(size="small"):
    """Generate extra metadata of varying sizes."""
    if size == "none":
        return {}
    elif size == "small":
        return {
            "source": "test_script",
            "version": "1.0",
            "tags": ["test", "generated"],
        }
    elif size == "medium":
        return {
            "source": "test_script",
            "version": "1.0",
            "tags": ["test", "generated", "medium"],
            "metadata": {
                "created_by": "test_user",
                "project": "asset_testing",
                "environment": "development",
            },
            "description": "A" * 200,  # 200 character description
        }
    elif size == "large":
        return {
            "source": "test_script",
            "version": "1.0",
            "tags": ["test", "generated", "large"] + [f"tag_{i}" for i in range(20)],
            "metadata": {
                "created_by": "test_user",
                "project": "asset_testing",
                "environment": "development",
                "config": {f"key_{i}": f"value_{i}" for i in range(50)},
            },
            "description": "B" * 500,  # 500 character description
            "notes": "C" * 300,
        }
    return {}


def generate_asset_payload(extra_size="small"):
    """Generate a single asset payload."""
    return {
        "checksum": generate_hex_checksum(),
        "display": generate_display(),
        "extra": generate_extra(extra_size),
    }


def generate_batch_payload(num_assets=1000, extra_size="small"):
    """Generate a batch of asset payloads."""
    return {"assets": [generate_asset_payload(extra_size) for _ in range(num_assets)]}


def format_size(size_bytes):
    """Format bytes to human-readable size."""
    for unit in ["B", "KB", "MB", "GB"]:
        if size_bytes < 1024.0:
            return f"{size_bytes:.2f} {unit}"
        size_bytes /= 1024.0
    return f"{size_bytes:.2f} TB"


def main():
    num_assets = int(sys.argv[1]) if len(sys.argv) > 1 else 1000
    extra_size = sys.argv[2] if len(sys.argv) > 2 else "small"

    if extra_size not in ["none", "small", "medium", "large"]:
        print(f"Invalid extra_size: {extra_size}")
        print("Valid options: none, small, medium, large")
        sys.exit(1)

    print(f"Generating {num_assets} assets with '{extra_size}' extra data...")

    payload = generate_batch_payload(num_assets, extra_size)
    json_str = json.dumps(payload, indent=2)

    # Write to file
    filename = f"assets_{num_assets}_{extra_size}.json"
    with open(filename, "w") as f:
        f.write(json_str)

    size = len(json_str.encode("utf-8"))
    print(f"✓ Generated: {filename}")
    print(f"✓ Size: {format_size(size)}")
    print(f"✓ Assets: {num_assets}")
    print("\nTo test with curl:")
    print("curl -X POST http://localhost:8080/api/assets \\")
    print('  -H "Content-Type: application/json" \\')
    print(f"  -d @{filename}")


if __name__ == "__main__":
    main()
