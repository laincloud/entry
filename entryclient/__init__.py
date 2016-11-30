#!/usr/bin/env python
# -*- coding: utf-8 -*-

import codecs
import errno
import fcntl
import message_pb2
import os
import platform
import select
import signal
import ssl
import struct
import sys
import termios
import tty
import websocket


class EntryClient:

    def __init__(self, endpoint, header=None):
        UTF8Reader = codecs.getreader('utf8')
        self._utf_in = UTF8Reader(sys.stdin)
        UTF8Writer = codecs.getwriter('utf8')
        self._utf_out = UTF8Writer(sys.stdout)
        self._utf_err = UTF8Writer(sys.stderr)

        self._oldtty = termios.tcgetattr(self._utf_in)
        self._old_handler = signal.getsignal(signal.SIGWINCH)
        try:
            sslopt = {"cert_reqs": ssl.CERT_NONE}
            self._ws = websocket.create_connection(
                url=endpoint, header=header, sslopt=sslopt)
        except:
            raise

    def invoke_shell(self):

        def on_term_resize(signum, frame):
            self._send_window_resize()
        signal.signal(signal.SIGWINCH, on_term_resize)

        try:
            tty.setraw(self._utf_in.fileno())
            tty.setcbreak(self._utf_in.fileno())
            self._send_window_resize()
            read_list = [self._ws.sock, self._utf_in]

            while True:
                try:
                    r, w, e = select.select(read_list, [], [])
                    if self._ws.sock in r:
                        data = self._ws.recv()
                        if self._is_close_message(data):
                            break
                    if self._utf_in in r:
                        utf8char = ""
                        valid_utf8 = False
                        while not valid_utf8:
                            utf8char = utf8char + os.read(
                                self._utf_in.fileno(), 1)
                            try:
                                utf8char.decode('utf-8')
                                valid_utf8 = True
                            except UnicodeDecodeError:
                                pass
                        self._ws.send(self._gen_plain_request(utf8char))
                except (select.error, IOError) as e:
                    if e.args and e.args[0] == errno.EINTR:
                        pass
                    else:
                        raise
        except websocket.WebSocketException:
            raise
        finally:
            self._close()

    def attach_container(self):
        try:
            while True:
                data = self._ws.recv()
                if not data or self._is_close_message(data):
                    break
        except:
            raise
        finally:
            self._close()

    def _close(self):
        termios.tcsetattr(self._utf_in, termios.TCSADRAIN, self._oldtty)
        signal.signal(signal.SIGWINCH, self._old_handler)
        self._ws.close()
        print ''

    def _is_close_message(self, msg):
        resp_msg = self._gen_response(msg)
        is_close = resp_msg.msgType == message_pb2.ResponseMessage.CLOSE
        if resp_msg.msgType == message_pb2.ResponseMessage.STDOUT:
            self._utf_out.write(resp_msg.content.decode('utf-8', 'replace'))
            self._utf_out.flush()
        elif (resp_msg.msgType == message_pb2.ResponseMessage.STDERR
              or resp_msg.msgType == message_pb2.ResponseMessage.CLOSE):
            self._utf_err.write(resp_msg.content.decode('utf-8', 'replace'))
            self._utf_err.flush()
        return is_close

    def _gen_resize_request(self, width, height):
        req_message = message_pb2.RequestMessage()
        req_message.msgType = message_pb2.RequestMessage.WINCH
        req_message.content = "%d %d" % (width, height)
        return req_message.SerializeToString()

    def _gen_plain_request(self, content):
        req_message = message_pb2.RequestMessage()
        req_message.msgType = message_pb2.RequestMessage.PLAIN
        req_message.content = content
        return req_message.SerializeToString()

    def _gen_response(self, payload):
        resp_message = message_pb2.ResponseMessage()
        resp_message.ParseFromString(payload)
        return resp_message

    def _get_window_size(self, utf_out):
        width, height = 80, 24
        # Can't do much for Windows
        if platform.system() != 'Windows':
            fmt = 'HH'
            result = fcntl.ioctl(
                utf_out.fileno(), termios.TIOCGWINSZ, struct.pack(fmt, 0, 0))
            height, width = struct.unpack(fmt, result)
        return width, height

    def _send_window_resize(self):
        width, height = self._get_window_size(self._utf_out)
        self._ws.send(self._gen_resize_request(width, height))
