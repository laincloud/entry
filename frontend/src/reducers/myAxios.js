import axios from 'axios'

const myAxios = axios.create({
  headers: {
    'Accept': 'application/vnd.laincloud.entry.v3+json'
  }
})

export const LIMIT = 200

export const get = (url, f, config = {}) => {
  return myAxios.get(url, config)
    .then(f)
    .catch(err => {
      if (err.response.status === 401) {
        window.alert('Please login with an account who own Entry.')
        return
      }

      console.error('error', err.response)
    })
}
