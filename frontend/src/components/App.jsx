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

import Commands from '../containers/Commands';
import Sessions from '../containers/Sessions';
import SessionReplay from '../containers/SessionReplay';
import {
  get
} from '../reducers/myAxios'

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
    marginTop: theme.spacing.unit * 10,
    marginBottom: theme.spacing.unit * 10
  }
});

class App extends React.Component {
  state = {
    user: window.localStorage.getItem('user'),
  };

  componentDidMount = () => {
    let params = new URLSearchParams(window.location.search.substring(1));
    let user = params.get('user');
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

  render() {
    const {
      classes,
      commandsCount,
      commandsRowsPerPage,
      sessionsCount,
      sessionsRowsPerPage,
      tabIndex,
      onChangeTabIndex,
    } = this.props;
    const {
      user,
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

            <Tabs value={tabIndex} onChange={onChangeTabIndex}>
              <Tab label="Sessions" />
              <Tab label="Commands" />
              {tabIndex === 2 && <Tab label="Replay" />}
            </Tabs>
          </AppBar>

          <div style={{display: (tabIndex === 0) ? 'block' : 'none'}}>
            <TabContainer>
              <div className={classes.main}>
                <Sessions count={sessionsCount} rowsPerPage={sessionsRowsPerPage}/>
              </div>
            </TabContainer>
          </div>

          <div style={{display: (tabIndex === 1) ? 'block' : 'none'}}>
            <TabContainer>
              <div className={classes.main}>
                <Commands count={commandsCount} rowsPerPage={commandsRowsPerPage}/>
              </div>
            </TabContainer>
          </div>

          {tabIndex === 2 &&
          <div>
            <TabContainer>
              <div className={classes.main}>
                <SessionReplay />
              </div>
            </TabContainer>
          </div>
          }
        </div>
      </MuiPickersUtilsProvider>
    );
  };
};

App.propTypes = {
  classes: PropTypes.object.isRequired,
  commandsCount: PropTypes.number.isRequired,
  commandsRowsPerPage: PropTypes.number.isRequired,
  sessionsCount: PropTypes.number.isRequired,
  sessionsRowsPerPage: PropTypes.number.isRequired,
  tabIndex: PropTypes.number.isRequired,
  onChangeTabIndex: PropTypes.func.isRequired
};

export default withStyles(styles)(App);
