"""Entrypoint for cli, enables execution with `python -m galaxy_update`."""

from galaxy_update.cli import cli

if __name__ == "__main__":
    cli()
