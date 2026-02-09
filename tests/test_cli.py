from click.testing import CliRunner

from galaxy_update.__main__ import cli

REQUIREMENTS_YML = """---
collections:
  - name: ansible.utils
    version: 1.0.0
"""

API_RESPONSE = {
    "highest_version": {
        "href": "/api/v3/plugin/ansible/content/published/collections/index/ansible/utils/versions/6.0.0/",
        "version": "6.0.0",
    },
}


def test_cli_updates_version(tmp_path, mocker):
    req_path = tmp_path / "requirements.yml"
    req_path.write_text(REQUIREMENTS_YML)

    mock_response = mocker.Mock()
    mock_response.raise_for_status.return_value = None
    mock_response.json.return_value = API_RESPONSE
    mocker.patch("httpx.AsyncClient.get", return_value=mock_response)

    runner = CliRunner()
    result = runner.invoke(cli, [str(req_path)])
    assert result.exit_code == 0
    out = req_path.read_text()
    assert "version: 6.0.0" in out


def test_cli_no_args():
    runner = CliRunner()
    result = runner.invoke(cli, [])
    assert result.exit_code != 0
    assert "Missing argument" in result.output


def test_cli_file_not_found():
    runner = CliRunner()
    result = runner.invoke(cli, ["nonexistent.yml"])
    assert result.exit_code != 0
