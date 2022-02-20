import React, { Component } from 'react';
import $ from 'jquery'
import './Profile.css'
import avatar from './images/img_avatar.png'
import Modal from 'react-modal'

function createAudio(src, parent) {
    const source = $(document.createElement('source'))
    source.attr('src', src)
    source.attr('type', 'audio/x-m4a')

    const audio = $(document.createElement('audio'))
    audio.attr('controls','')
    audio.append(source)
    parent.append(audio)
    return audio
}

function playEpisode(episdoe, callback) {
    $.getJSON('/episodes/one?name=' + episdoe, function (item){
        createAudio(item.spec.audioSource, $('#global-audio-zone').show()).trigger('play').on('ended', function () {
            const episode = $(this).attr('episode')
            const profile = window.localStorage.getItem('profile')
            $.post('/profile/playOver?name=' + profile + '&episode=' + episode, function (){
                $('span[episode=' + episode + ']').remove()

                if (callback) {
                    callback()
                }
            })
        }).attr('episode', item.metadata.name)
    })
}

const customStyles = {
    content: {
        top: '100px',
        right: '100px',
        bottom: '80px',
    },
};
Modal.setAppElement('#root');
class ProfileModal extends Component {
    constructor(props) {
        super(props);
        this.state = {
            laterPlayList: [],
            isOpen: false,
            rssURL: "",
            rssName: "",
            github: ""
        }
    }

    closeModal() {
        this.setState({isOpen: false})
    }

    addRSS() {
        const requestOptions = {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({
                address: this.state.rssURL,
                name: this.state.rssName
            })
        };
        fetch('/rsses', requestOptions)
    }
    setRSSURL(e) {
        this.setState({
            rssURL: e.target.value
        })
    }
    setRSSName(e) {
        this.setState({
            rssName: e.target.value
        })
    }

    reload(){
        const name = localStorage.getItem('profile')
        if (name === "" || !name || name == null) {
            return
        }
        const profileObj = this
        fetch('/profiles?name=' + name)
            .then(res => res.json())
            .then(res => {
                this.setState({
                    laterPlayList: res.spec.laterPlayList
                })

                if (res.spec.socialLinks) {
                    const github = res.spec.socialLinks["github"]
                    this.setState({
                        github: github
                    })
                    if (github !== "") {
                        $('#avatar').attr('src', 'https://avatars.githubusercontent.com/' + github)
                    }
                }
            })
    }

    onOpen() {
        $('#social-account-github').val(this.state.github)
    }

    componentDidMount() {
        this.reload()
        const com = this
        $('#profiles').on('reload', function (){
            com.reload()
        })
    }

    setGitHubAccount(currentValue) {
        const oldValue = this.state.github
        if (currentValue === oldValue) {
            return
        }
        const profile = window.localStorage.getItem('profile')
        const requestOptions = {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' }
        };
        fetch('/profile/social?kind=github&account=' + currentValue + '&name=' + profile, requestOptions)
            .then(() => (
                this.setState({
                    github: currentValue
                })
            ))
    }

    play() {
        $('span[episode]').each(function (i, item){
            item = $(item)
            const episode = item.attr('episode')
            $('#' + episode).trigger('play-audio').on('ended-audio', function () {
                item.remove()
                // playButton.click()
            })

            // hideGlobalAudioZone()
            if ($('#' + episode).length > 0) {
                $([document.documentElement, document.body]).animate({
                    scrollTop: $('#' + episode).offset().top
                }, 2000);
            } else {
                playEpisode(episode, function () {
                })
            }
            return false
        })
    }

    render() {
        const {laterPlayList} = this.state;
        return (
            <div>
                <Modal
                    isOpen={this.state.isOpen}
                    onAfterOpen={() => this.onOpen()}
                    contentLabel="Example Modal"
                    style={customStyles}
                >
                    <button onClick={() => this.closeModal()} className="modal-close-but">Close</button>

                    <div style={{display: "none"}} id="login-zone">
                        <label>
                            Name: <input name="name" id="login-name" />
                        </label>
                        <div><button action="login">Login</button></div>
                        <div><button action="register">Register</button></div>
                    </div>
                    <div>
                        <div>
                            <span>Listen Later List: </span>
                            <button onClick={this.play}>Play</button>
                            {laterPlayList.map((item, index) => (
                                    <span episode={item.name} key={index} className="later-play-item">{item.displayName}</span>
                                )
                            )}
                        </div>
                    </div>

                    <div>
                        New RSS feed:<input onChange={(e) => this.setRSSURL(e)}/> with name:
                        <input onChange={(e) => this.setRSSName(e)}/><button onClick={() => this.addRSS()}>Add</button>
                    </div>

                    <div className="social-account-zone">
                        <div>Social Account</div>
                        <div>GitHub: <input id="social-account-github" onBlur={(e) => this.setGitHubAccount(e.target.value)}/></div>
                    </div>
                </Modal>
            </div>
        );
    }
}

class Profile extends Component {
    constructor(props) {
        super(props);
        this.state = {
            showModal: false,
        };
        this.profileModalElement = React.createRef();
    }

    toggleModal() {
        this.profileModalElement.current.setState({isOpen: true})
    }

    render() {
        return (
            <div id="profiles">
                <img src={avatar} className="avatar" id="avatar" alt="Avatar" onClick={() => this.toggleModal()}/>

                <ProfileModal ref={this.profileModalElement}/>
            </div>
        );
    }
}

export default Profile;
