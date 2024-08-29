"""Update dependencies in requirements.yml."""

import asyncio
from pathlib import Path

import click
import httpx
import yaml


@click.command()
@click.argument(
    "requirements",
    nargs=-1,
    required=True,
    type=click.Path(exists=True, dir_okay=False, path_type=Path),
)
def cli(requirements: tuple[Path]) -> None:
    """Update dependencies in requirements.yml."""
    asyncio.run(update_requirements(requirements))


async def update_requirements(requirements: tuple[Path]) -> None:
    """Update dependencies in requirements.yml."""
    # Create a HTTP client
    async with httpx.AsyncClient() as client:
        base_url = "https://galaxy.ansible.com/api"
        collections_index = "v3/plugin/ansible/content/published/collections/index"

        for requirements_file in requirements:
            # Load the requirements file
            reqs = yaml.safe_load(requirements_file.read_text())

            # Update the versions
            for collection in reqs["collections"]:
                name = collection["name"].replace(".", "/")
                response = await client.get(f"{base_url}/{collections_index}/{name}/versions/")
                response.raise_for_status()
                collection["version"] = response.json()["data"][0]["version"]

            # Write the updated requirements back to the file
            requirements_file.write_text(yaml.safe_dump(reqs, explicit_start=True))
