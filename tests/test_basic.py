def test_sanity():
    assert True, 'Basic sanity check passes'

def test_environment():
    import os
    assert os.environ.get('ENV', 'dev') == 'dev', 'Environment is dev'
