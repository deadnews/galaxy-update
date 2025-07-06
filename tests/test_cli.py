from click.testing import CliRunner

from galaxy_update.__main__ import cli

REQUIREMENTS_YML = """---
collections:
  - name: ansible.utils
    version: 1.0.0
"""

API_RESPONSE = {
    "meta": {"count": 58},
    "links": {},
    "data": [
        {
            "version": "6.0.0",
            "href": "/api/v3/plugin/ansible/content/published/collections/index/ansible/utils/versions/6.0.0/",
            "created_at": "2025-04-14T06:20:22.435485Z",
            "updated_at": "2025-04-14T06:20:22.587797Z",
            "requires_ansible": ">=2.16.0",
            "marks": [],
        },
    ],
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
