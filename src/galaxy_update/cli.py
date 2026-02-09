"""Update dependencies in requirements.yml."""

import asyncio
from pathlib import Path

import click
import httpx
import pyaml
import yaml

GALAXY_API = "https://galaxy.ansible.com/api"
COLLECTIONS_INDEX = "v3/plugin/ansible/content/published/collections/index"


@click.command()
@click.argument(
    "requirements",
    nargs=-1,
    required=True,
    type=click.Path(exists=True, dir_okay=False, path_type=Path),
)
@click.version_option()
def cli(requirements: tuple[Path]) -> None:
    """Update dependencies in requirements.yml."""
    asyncio.run(update_requirements(requirements))


async def fetch_latest_version(client: httpx.AsyncClient, name: str) -> str:
    """Fetch the highest version of a collection from Ansible Galaxy."""
    namespace, collection = name.split(".", maxsplit=1)
    url = f"{GALAXY_API}/{COLLECTIONS_INDEX}/{namespace}/{collection}/"
    response = await client.get(url)
    response.raise_for_status()

    return response.json()["highest_version"]["version"]


async def update_requirements(requirements: tuple[Path]) -> None:
    """Update dependencies in requirements.yml."""
    async with httpx.AsyncClient() as client:
        for requirements_file in requirements:
            reqs = yaml.safe_load(requirements_file.read_text())
            collections = reqs["collections"]

            versions = await asyncio.gather(
                *(fetch_latest_version(client, c["name"]) for c in collections)
            )
            for collection, version in zip(collections, versions, strict=True):
                collection["version"] = version

            requirements_file.write_text(str(pyaml.dump(reqs, explicit_start=True)))
