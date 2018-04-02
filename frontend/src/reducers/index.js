import {
  combineReducers
} from 'redux'

import app from './app'
import commands from './commands'
import sessions from './sessions'

export default combineReducers({
  app,
  commands,
  sessions
})
