#!/usr/bin/env python3

import os
import sys
import json
import http
import urllib.request
import urllib.parse

http.client.HTTPConnection.debuglevel = 1

owner = 'heraldgo'
repo = 'heraldd'

url_releases = 'https://api.github.com/repos/{owner}/{repo}/releases'.format(
    owner=owner, repo=repo)

github_token = os.environ.get('GITHUB_TOKEN')


def get_upload_url(version):
    release = 'v' + version

    try:
        resp = urllib.request.urlopen(
            '{0}/tags/{1}'.format(url_releases, release))
        if resp.getcode() == 200:
            return json.loads(resp.read()).get('upload_url', '')
    except urllib.error.HTTPError as e:
        if e.code != 404:
            print('Get release HTTP error', e.code, e.reason, file=sys.stderr)
            return ''
    except OSError as e:
        print('Get release error:', e, file=sys.stderr)
        return ''

    print('Create new release for version', version)

    new_release = {
        'tag_name': release,
        'name': release,
        'body': 'Version '+version,
    }

    data = json.dumps(new_release).encode()
    create_request = urllib.request.Request(url_releases, data=data)
    create_request.add_header('Authorization', 'token '+github_token)

    try:
        resp = urllib.request.urlopen(create_request)
        if resp.getcode() != 201:
            print('Failed to create release', file=sys.stderr)
            return ''
    except urllib.error.HTTPError as e:
        print('Create release HTTP error', e.code, e.reason, file=sys.stderr)
        return ''
    except OSError as e:
        print('Create release error:', e, file=sys.stderr)
        return ''

    upload_url = json.loads(resp.read()).get('upload_url', '')

    return upload_url


def upload_asset(url_upload, file_path):
    file_name = os.path.basename(file_path)

    with open(file_path, 'rb') as f:
        data = f.read()

    upload_request = urllib.request.Request(
        '{0}?name={1}'.format(url_upload, file_name), data=data)
    upload_request.add_header('Authorization', 'token '+github_token)
    upload_request.add_header('Content-Type', 'application/zip')

    try:
        resp = urllib.request.urlopen(upload_request)
        if resp.getcode() != 201:
            print('Failed to upload asset', file=sys.stderr)
            return
    except urllib.error.HTTPError as e:
        print('Upload asset HTTP error', e.code, e.reason, file=sys.stderr)
        return
    except OSError as e:
        print('Upload asset error:', e, file=sys.stderr)
        return

    print('Upload file successfully:', file_path)


def main():
    if len(sys.argv) < 2:
        print('Not enough arguments', file=sys.stderr)
        return 1

    version = sys.argv[1]
    assets = sys.argv[2:]

    if not github_token:
        print('Github token not provided', file=sys.stderr)
        return 1

    upload_url = get_upload_url(version)
    if not upload_url:
        print('Upload url not found', file=sys.stderr)
        return 1

    url_upload = upload_url.split('{', 1)[0]
    print("URL upload:", url_upload)

    for f in assets:
        upload_asset(url_upload, f)

    return 0


if __name__ == '__main__':
    sys.exit(main())
