import React from 'react';
import {
  Terminal
} from 'xterm';
import 'xterm/dist/xterm.css';
import {
  withStyles
} from 'material-ui/styles';
import PropTypes from 'prop-types';
import Card, {
  CardHeader,
  CardContent,
  CardActions
} from 'material-ui/Card';
import IconButton from 'material-ui/IconButton';
import ReplayIcon from 'material-ui-icons/Replay';
import * as fit from 'xterm/lib/addons/fit/fit';

import './SessionReplay.css';

// The following constants must be equals with
// https://github.com/laincloud/entry/blob/master/message.proto
const RESPONSE_CLOSE = 2;
const RESPONSE_PING = 3;

Terminal.applyAddon(fit);

var printInfo = (data) => {
  return '\x1B[32m>>> ' + data + '\x1B[0m';
}

var printError = (data) => {
  return '\x1B[31m>>> ' + data + '\x1B[0m';
}

const styles = {
  content: {
    height: '50vh',
    margin: 'auto'
  },
  replay: {
    marginLeft: 'auto'
  }
};

class SessionReplay extends React.Component {
  state = {
    term: null,
    ws: null
  }

  componentDidMount = () => {
    let term = new Terminal({
      cursorBlink: true
    });
    term.open(document.getElementById('term'));
    term.fit();
    this.setState({
      term: term
    }, this.handleReplay);
  }

  componentWillUnmount = () => {
    const {
      term,
      ws
    } = this.state;

    term.destroy();
    if (ws) {
      ws.close()
    };
    this.setState({
      term: null,
      ws: null
    });
  }

  handleReplay = () => {
    const {
      term
    } = this.state;
    const {
      sessionID
    } = this.props;

    if (this.state.ws) {
      this.state.ws.close();
    };
    term.reset();
    term.focus();

    let replayURI = 'wss://' + window.location.host + '/api/sessions/';
    replayURI += sessionID + '/replay';
    let ws = new WebSocket(replayURI);
    this.setState({
      ws: ws
    });

    ws.onopen = () => {
      console.info('WebSocket is open...');
      term.writeln(printInfo('Session replay started...'));
    };

    ws.onclose = () => {
      console.info('WebSocket has been closed.');
    }

    ws.onerror = (event) => {
      console.error('WebSocket.onerror(%s)', event);
      term.writeln(printError(
        'Server stops the connection. Please ask admin for help.'));
    };

    ws.onmessage = (message) => {
      let reader = new FileReader();
      reader.onloadend = () => {
        let decoder = new TextDecoder('utf-8');
        let data = JSON.parse(decoder.decode(reader.result));

        if (data.msgType === RESPONSE_PING) {
          console.info('Websocket.onmessage(ping)');
          return
        }

        let msg = atob(data.content);
        if (msg.charAt(msg.length - 1) === '\n') {
          msg += '\r';
        }
        term.write(msg);

        if (data.msgType === RESPONSE_CLOSE) {
          ws.close();
          return
        };
      };
      reader.readAsArrayBuffer(message.data);
    };

    window.onresize = () => {
      term.fit();
    };

    window.onbeforeunload = () => {
      ws.close();
    };
  }

  render() {
    const {
      classes,
      sessionID
    } = this.props;

    return (
      <div>
        <Card className={classes.card}
        >
          <CardHeader
            title={'Replay Session: ' + sessionID}
          >
          </CardHeader>

          <CardContent
            id="term"
            className={classes.content}
          >
          </CardContent>

          <CardActions
          >
            <IconButton
              className={classes.replay}
              aria-label="Replay"
              onClick={this.handleReplay}
            >
              <ReplayIcon />
            </IconButton>
          </CardActions>
        </Card>
      </div>
    )
  }
}

SessionReplay.propTypes = {
  classes: PropTypes.object.isRequired,
  sessionID: PropTypes.number.isRequired,
}

export default withStyles(styles)(SessionReplay);
