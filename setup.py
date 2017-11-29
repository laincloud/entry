from setuptools import setup, find_packages

requirements = [
    'websocket-client==0.32.0',
    'protobuf==3.0.0b3',
    'six>=1.9.0',
]

setup(
    name="entryclient",
    author='EricPai <ericpai94@hotmail.com>',
    version='2.3.2',
    packages=find_packages(),
    include_package_data=True,
    install_requires=requirements,
)
