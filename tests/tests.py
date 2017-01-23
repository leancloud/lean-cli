#! /usr/bin/env python3

import os
import json
import tempfile


ACCOUNT = os.environ['TEST_ACCOUNT']
PASSWORD = os.environ['TEST_PASSWORD']
COOKIES_PATH = os.path.expanduser('~/.config/leancloud/cookies')


def test_basic():
    try:
        os.remove(os.path.expanduser(COOKIES_PATH))
    except FileNotFoundError:
        pass
    assert os.system('lean') == 0


def test_login():
    assert os.system('lean login %s %s' % (ACCOUNT, PASSWORD)) == 0
    with open(COOKIES_PATH) as f:
        json.loads(f.read())


def test_deploy_from_git():
    tempdir = tempfile.mkdtemp()
    print(tempdir)
    os.chdir(tempdir)
    assert os.system('lean checkout uwju5zvj7d0uumrwogfl78x98hxg3bnyjhdznenxukhujiz7') == 0
    assert os.system('lean deploy -g') == 0
