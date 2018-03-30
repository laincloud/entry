import {
  connect
} from 'react-redux'

import SessionReplay from '../components/SessionReplay.jsx'

const mapStateToProps = state => ({
  sessionID: state.app.replaySessionID
})

export default connect(mapStateToProps)(SessionReplay)
