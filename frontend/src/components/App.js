import React from 'react';
import PropTypes from 'prop-types';
import {
  withStyles
} from 'material-ui/styles';
import AppBar from 'material-ui/AppBar';
import Toolbar from 'material-ui/Toolbar';
import Typography from 'material-ui/Typography';
import Button from 'material-ui/Button';
import Tabs, {
  Tab
} from 'material-ui/Tabs';
import {
  MuiPickersUtilsProvider
} from 'material-ui-pickers';
import DateFnsUtils from 'material-ui-pickers/utils/date-fns-utils';

import Commands from './Commands';
import Sessions from './Sessions';
import {
  get
} from '../MyAxios.jsx';
import './App.css';

function TabContainer(props) {
  return (
    <Typography component="div" style={{ padding: 8 * 3}}>
      {props.children}
    </Typography>
  );
}

TabContainer.propTypes = {
  children: PropTypes.node.isRequired,
};

const styles = theme => ({
  root: {
    flexGrow: 1,
    backgroundColor: theme.palette.background.paper
  },
  flex: {
    flex: 1
  },
  floatLeft: {
    float: 'left',
    marginRight: theme.spacing.unit * 2
  },
  center: {
    display: 'flex',
    alignItems: 'center'
  },
  main: {
    width: '90%',
    marginLeft: 'auto',
    marginRight: 'auto',
    marginTop: theme.spacing.unit * 15,
    marginBottom: theme.spacing.unit * 15
  }
});

class App extends React.Component {
  state = {
    user: window.localStorage.getItem('user'),
    commandsSessionID: '',
    commandsSince: new Date(),
    tabIndex: 0,
    sessionsTabStyle: {
      display: 'block'
    },
    commandsTabStyle: {
      display: 'none'
    }
  };

  componentDidMount = () => {
    let params = new URLSearchParams(window.location.search.substring(1));
    let user = params.get("user");
    if (user) {
      this.setState({
        user: user
      });
      window.localStorage.setItem('user', user);
    }
  };

  login = response => {
    let ssoConfig = response.data.sso;
    let url = 'https://' + ssoConfig.domain + '/oauth2/auth';
    let params = new URLSearchParams();
    params.append('response_type', 'code');
    params.append('client_id', ssoConfig.client_id);
    params.append('redirect_uri', ssoConfig.redirect_uri);
    params.append('scope', ssoConfig.scope);
    window.location.href = url + '?' + params.toString();
  };

  handleLogin = () => {
    get('/api/config', this.login);
  };

  logout = response => {
    this.setState({
      user: ''
    });
    window.localStorage.removeItem('user');
    window.location.href = '/web';
  };

  handleLogout = () => {
    get('/api/logout', this.logout);
  };

  handleSessionSearchCommands = (sessionID, since) => {
    this.setState({
      tabIndex: 1,
      sessionsTabStyle: {
        display: 'none'
      },
      commandsTabStyle: {
        display: 'block'
      },
      commandsSessionID: sessionID,
      commandsSince: since
    });
  };

  handleCommandsSessionIDChange = event => {
    this.setState({
      commandsSessionID: event.target.value
    })
  }

  handleCommandsSinceChange = since => {
    this.setState({
      comandsSince: since
    })
  }

  handleTabIndexChange = (event, tabIndex) => {
    if (tabIndex === 0) {
      this.setState({
        tabIndex: tabIndex,
        sessionsTabStyle: {
          display: 'block'
        },
        commandsTabStyle: {
          display: 'none'
        }
      });
    } else {
      this.setState({
        tabIndex: tabIndex,
        sessionsTabStyle: {
          display: 'none'
        },
        commandsTabStyle: {
          display: 'block'
        }
      });
    }
  };

  render() {
    const {
      classes
    } = this.props;
    const {
      user,
      tabIndex,
      commandsSessionID,
      commandsSince,
      sessionsTabStyle,
      commandsTabStyle
    } = this.state;

    return (
      <MuiPickersUtilsProvider utils={DateFnsUtils}>
        <div className={classes.root}>
          <AppBar position="static">
            <Toolbar>
              <Typography variant="title" color="inherit" className={classes.flex}>
                Entry
              </Typography>

              {!user && (<Button color="inherit" onClick={this.handleLogin}>Login</Button>)}
              {user && (
                <div className={classes.center}>
                  <Typography
                    paragraph={false}
                    variant="subheading"
                    color="inherit"
                    className={classes.floatLeft}
                  >
                    {user}
                  </Typography>

                  <Button color="inherit" onClick={this.handleLogout}>Logout</Button>
                </div>
              )}
            </Toolbar>

            <Tabs value={tabIndex} onChange={this.handleTabIndexChange}>
              <Tab label="Sessions" />
              <Tab label="Commands" />
            </Tabs>
          </AppBar>

          <div style={sessionsTabStyle}>
            <TabContainer>
              <div className={classes.main}>
                <Sessions onSearchCommands={this.handleSessionSearchCommands}></Sessions>
              </div>
            </TabContainer>
          </div>

          <div style={commandsTabStyle}>
            <TabContainer>
              <div className={classes.main}>
                <Commands
                  sessionID={commandsSessionID}
                  onSessionIDChange={this.handleCommandsSessionIDChange}
                  since={commandsSince}
                  onSinceChange={this.handleCommandsSinceChange}
                >
                </Commands>
              </div>
            </TabContainer>
          </div>
        </div>
      </MuiPickersUtilsProvider>
    );
  };
};

App.propTypes = {
  classes: PropTypes.object.isRequired
};

export default withStyles(styles)(App);
