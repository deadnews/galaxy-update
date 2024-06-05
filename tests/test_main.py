from galaxy_update.__main__ import cli


def test_main() -> None:
    assert callable(cli)
