import React from 'react';
import Button from 'material-ui/Button'
import PropTypes from 'prop-types'
import TextField from 'material-ui/TextField';
import {
  withStyles
} from 'material-ui/styles'

import MyDateTimePicker from './MyDateTimePicker.jsx';

const styles = theme => ({
  container: {
    display: 'flex',
    flexWrap: 'wrap',
    marginBottom: theme.spacing.unit * 10,
    justifyContent: 'center'
  },
  textField: {
    marginLeft: theme.spacing.unit,
    marginRight: theme.spacing.unit,
    minWidth: 200
  },
  menu: {
    width: 200
  },
  button: {
    margin: theme.spacing.unit * 3,
  }
})

const CommandsQuery = ({
  classes,
  appName,
  content,
  sessionID,
  since,
  user,
  onClick,
  onSinceChange,
  onChangeTextField
}) => (
  <form className={classes.container}>
    <TextField
      id="user"
      label="User"
      className={classes.textField}
      value={user}
      onChange={event => onChangeTextField('user', event.target.value)}
      margin="normal"
    />

    <TextField
      id="appName"
      label="App Name"
      className={classes.textField}
      value={appName}
      onChange={event => onChangeTextField('appName', event.target.value)}
      margin="normal"
    />

    <TextField
      id="content"
      label="Content(MySQL LIKE)"
      className={classes.textField}
      value={content}
      onChange={event => onChangeTextField('content', event.target.value)}
      margin="normal"
    />

    <TextField
      id="sessionID"
      label="Session ID"
      className={classes.textField}
      value={sessionID}
      onChange={event => onChangeTextField('sessionID', event.target.value)}
      type="number"
      InputLabelProps={{
        shrink: true
      }}
      margin="normal"
    />

    <MyDateTimePicker
      value={since}
      onChange={since => onChangeTextField('since', since)}
    />

    <Button
      variant="raised"
      color="primary"
      className={classes.button}
      onClick={onClick}
    >
      Query
    </Button>
  </form>
);

CommandsQuery.propTypes = {
  classes: PropTypes.object.isRequired,
  appName: PropTypes.string.isRequired,
  content: PropTypes.string.isRequired,
  sessionID: PropTypes.string.isRequired,
  since: PropTypes.object.isRequired,
  user: PropTypes.string.isRequired,
  onClick: PropTypes.func.isRequired,
  onChangeTextField: PropTypes.func.isRequired
};

export default withStyles(styles)(CommandsQuery);
