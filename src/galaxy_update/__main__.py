"""Entrypoint for cli, enables execution with `python -m gitlab_template_python`."""

from galaxy_update.app import cli

if __name__ == "__main__":
    cli()
