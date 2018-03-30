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

const CommandsTableToolbar = ({
  classes
}) => (
  <Toolbar className={classNames(classes.root)}>
    <div className={classes.title}>
      <Typography variant="title">Commands</Typography>
    </div>
  </Toolbar>
)

CommandsTableToolbar.propTypes = {
  classes: PropTypes.object.isRequired
}

export default withStyles(styles)(CommandsTableToolbar)
