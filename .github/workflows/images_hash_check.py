import json
import requests
import sys
import argparse

parser = argparse.ArgumentParser(description="Validate image hashes against Docker Hub manifest digests")
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
        "Accept": "application/vnd.docker.distribution.manifest.list.v2+json"
    }
    r = requests.get(f"https://registry-1.docker.io/v2/{repo}/manifests/{tag}", headers=headers)
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

for version in data["versions"]:
    for product, versions in version["matrix"].items():
        for ver, details in versions.items():
            image = details["image_path"]

            expected_amd = details.get("image_hash")
            if expected_amd:
                print(f"🔎 Checking {image} (amd64) ...")
                try:
                    digest = get_digest(image, "amd64")
                    if digest == f"sha256:{expected_amd}":
                        print(f"✅ {image} (amd64) OK")
                    else:
                        print(f"❌ {image} (amd64) mismatch: expected sha256:{expected_amd}, got {digest}")
                        errors.append((image, "amd64", expected_amd, digest))
                except Exception as e:
                    print(f"⚠️ Failed to check {image} (amd64): {e}")
                    errors.append((image, "amd64", expected_amd, "ERROR"))

            expected_arm = details.get("image_hash_arm64")
            if expected_arm:
                print(f"🔎 Checking {image} (arm64) ...")
                try:
                    digest = get_digest(image, "arm64")
                    if digest == f"sha256:{expected_arm}":
                        print(f"✅ {image} (arm64) OK")
                    else:
                        print(f"❌ {image} (arm64) mismatch: expected sha256:{expected_arm}, got {digest}")
                        errors.append((image, "arm64", expected_arm, digest))
                except Exception as e:
                    print(f"⚠️ Failed to check {image} (arm64): {e}")
                    errors.append((image, "arm64", expected_arm, "ERROR"))

if errors:
    print("\n❌ Some images did not match:")
    for img, arch, exp, got in errors:
        print(f"- {img} ({arch}): expected sha256:{exp}, got {got}")
    sys.exit(1)
else:
    print("\n🎉 All image hashes match Docker Hub!")