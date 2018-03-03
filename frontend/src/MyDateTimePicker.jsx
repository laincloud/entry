import React, {
  Fragment
} from 'react';
import {
  DateTimePicker
} from 'material-ui-pickers';
import PropTypes from 'prop-types';

function MyDateTimePicker(props) {
  return (
    <Fragment>
      <div className="picker">
        <DateTimePicker
          value={props.value}
          onChange={props.onChange}
          ampm={false}
          disableFuture={true}
          format="YYYY-MM-DD hh:mm"
          label="Since"
          margin="normal"
        />
      </div>
    </Fragment>
  );
};

MyDateTimePicker.propTypes = {
  value: PropTypes.object.isRequired,
  onChange: PropTypes.func.isRequired
};

export default MyDateTimePicker;
