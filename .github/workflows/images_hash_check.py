import json
import requests
import sys
import argparse

parser = argparse.ArgumentParser(
    description="Validate image hashes against Docker Hub manifest digests"
)
parser.add_argument("-f", "--file", required=True, help="Path to the JSON file")
args = parser.parse_args()

json_file = args.file

with open(json_file) as f:
    data = json.load(f)


def get_digest(image, arch="amd64"):
    repo, tag = image.split(":")

    token_url = f"https://auth.docker.io/token?service=registry.docker.io&scope=repository:{repo}:pull"
    token = requests.get(token_url).json()["token"]

    headers = {
        "Authorization": f"Bearer {token}",
        "Accept": "application/vnd.docker.distribution.manifest.list.v2+json",
    }
    r = requests.get(
        f"https://registry-1.docker.io/v2/{repo}/manifests/{tag}", headers=headers
    )
    r.raise_for_status()
    manifest_list = r.json()

    if "manifests" in manifest_list:
        for m in manifest_list["manifests"]:
            if m["platform"]["architecture"] == arch:
                return m["digest"]
        raise ValueError(f"No manifest found for architecture {arch} in {image}")
    else:
        return r.headers["Docker-Content-Digest"]


errors = []


def check_hash(image, arch, expected_hash):
    try:
        digest = get_digest(image, arch)
        expected_digest = f"sha256:{expected_hash}"
        if digest == expected_digest:
            print(f"✅ {image} ({arch}) OK")
            return

        print(f"❌ {image} ({arch}) mismatch: expected {expected_digest}, got {digest}")
        errors.append((image, arch, expected_hash, digest))
    except Exception as e:
        print(f"⚠️ Failed to check {image} ({arch}): {e}")
        errors.append((image, arch, expected_hash, "ERROR"))


for version in data["versions"]:
    for product, versions in version["matrix"].items():
        for ver, details in versions.items():
            image = details["image_path"]

            hash_checks = [
                ("amd64", details.get("image_hash")),
                ("arm64", details.get("image_hash_arm64")),
            ]

            checked_any_hash = False
            for arch, expected_hash in hash_checks:
                if expected_hash is None:
                    continue

                checked_any_hash = True
                check_hash(image, arch, expected_hash)

            if not checked_any_hash:
                print(f"❌ {image} has no image_hash/image_hash_arm64")
                errors.append((image, "hash", "defined", "MISSING"))

if errors:
    print("\n❌ Some images did not match:")
    for img, arch, exp, got in errors:
        print(f"- {img} ({arch}): expected sha256:{exp}, got {got}")
    sys.exit(1)
else:
    print("\nAll image hashes match Docker Hub!")
