import React from 'react';
import Toolbar from 'material-ui/Toolbar';
import classNames from 'classnames'
import PropTypes from 'prop-types'
import {
  withStyles
} from 'material-ui/styles'
import Typography from 'material-ui/Typography'

const styles = theme => ({
  root: {
    paddingRight: theme.spacing.unit
  },
  title: {
    flex: '0 0 auto'
  }
})

const SessionsTableToolbar = ({
  classes
}) => (
  <Toolbar className={classNames(classes.root)}>
    <div className={classes.title}>
      <Typography variant="title">Sessions</Typography>
    </div>
  </Toolbar>
)

SessionsTableToolbar.propTypes = {
  classes: PropTypes.object.isRequired
}

export default withStyles(styles)(SessionsTableToolbar)
