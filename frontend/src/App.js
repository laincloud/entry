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
    value: 0,
    sessionsTabStyle: {
      display: 'block'
    },
    commandsTabStyle: {
      display: 'none'
    }
  };

  handleChange = (event, value) => {
    if (value === 0) {
      this.setState({
        value: value,
        sessionsTabStyle: {
          display: 'block'
        },
        commandsTabStyle: {
          display: 'none'
        }
      });
    } else {
      this.setState({
        value: value,
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
      value,
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

              <Button color="inherit">Login</Button>
            </Toolbar>

            <Tabs value={value} onChange={this.handleChange}>
              <Tab label="Sessions" />
              <Tab label="Commands" />
            </Tabs>
          </AppBar>

          <div style={sessionsTabStyle}>
            <TabContainer>
              <div className={classes.main}>
                <Sessions></Sessions>
              </div>
            </TabContainer>
          </div>

          <div style={commandsTabStyle}>
            <TabContainer>
              <div className={classes.main}>
                <Commands></Commands>
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
