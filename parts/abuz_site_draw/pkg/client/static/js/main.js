let currentPosition = 1

async function onTelegramAuth(user) {
    let t = [], ts = ""
    let user_new = {...user}
    delete user_new.hash
    Object.keys(user_new).map(e => {
        t.push(`${e}=${user_new[e]}`)
    })
    t.sort()
    ts = t.join("\n")
    const rawResponse = await fetch('/login', {
        method: 'POST',
        headers: {
            'Accept': 'application/json',
            'Content-Type': 'application/json'
        },
        body: JSON.stringify({telegram: user.username, hash: user.hash, hash_data: ts})
    });
    const content = await rawResponse.text();
    if (content === "OK") {
        window.location.href = "/lk"
    }
}

document.getElementById("horizontal-scroller")
    .addEventListener('wheel', function (event) {
        if (event.deltaMode === event.DOM_DELTA_PIXEL) {
            var modifier = 1;
            // иные режимы возможны в Firefox
        } else if (event.deltaMode === event.DOM_DELTA_LINE) {
            var modifier = parseInt(getComputedStyle(this).lineHeight);
        } else if (event.deltaMode === event.DOM_DELTA_PAGE) {
            var modifier = this.clientHeight;
        }
        if (event.deltaY !== 0) {
            // замена вертикальной прокрутки горизонтальной
            if (window.innerWidth <= 992) {
                if (event.deltaY < 0) currentPosition += 1;
                else currentPosition -= 1;

                if (currentPosition < 1) currentPosition = 5
                if (currentPosition > 5) currentPosition = 1

                this.scrollLeft = (currentPosition - 1) * this.offsetWidth
                console.log(currentPosition)
                for (let i = 1; i <= 5; i++) {
                    let item = document.getElementById(`scroll-ellips-${i}`)
                    if (i === currentPosition) {
                        item.classList.add("active")
                    } else {
                        item.classList.remove("active")
                    }
                }

            } else {
                this.scrollLeft += modifier * event.deltaY;
            }
            event.preventDefault();
        }
    });

function sleep(ms) {
    return new Promise(resolve => setTimeout(resolve, ms));
}

const ids = ["game-slot-1", "game-slot-7", "game-slot-5", "game-slot-2", "game-slot-4", "game-slot-8", "game-slot-3", "game-slot-6", "game-slot-4"]

document.getElementById("start").addEventListener('click', getPrice)

document.getElementById("start-mobile").addEventListener('click', getPrice)

async function getPrice(e) {
    const rawResponse = await fetch('/', {
        method: 'POST',
        headers: {
            'Accept': 'application/json',
            'Content-Type': 'application/json'
        },
        body: JSON.stringify({})
    });
    const content = await rawResponse.json();
    initializeClock("game-time", content.timer)
    initializeClock("game-time-mobile", content.timer)
    await sleep(500);
    for (let i = 0; i < ids.length; i++) {
        let item = document.getElementById(ids[i])
        item.classList.add("active");
        await sleep(500);
        item.classList.remove("active");
    }
    createModalPrice(content)
}

async function startRandomItems() {
    let tmp = ["flicker-item-1", "flicker-item-2", "flicker-item-3", "flicker-item-4", "flicker-item-5", "flicker-item-6"]
    while (tmp.length > 0) {
        var item = tmp[Math.floor(Math.random() * tmp.length)];
        tmp = tmp.filter(e => e !== item)
        let e = document.getElementById(item)
        e.classList.add("animate")
        await sleep(3000)
    }
}


function createModalPrice(data) {
    let typeD = document.getElementById("game-modal-type")
    let ticketD = document.getElementById("game-modal-ticket")
    let bodyD = document.getElementById("game-modal-body")
    let descriptionD = document.getElementById("game-modal-description")
    let takeD = document.getElementById("take-price")
    let copyD = document.getElementById("modal-copy")
    let toyD = document.getElementById("game-modal-toy")
    switch (data.price.type){
        case 0:
            ticketD.classList.add("green")
            typeD.innerText = "УВЫ :("
            bodyD.innerText = data.price.data
            bodyD.classList.add("none")
            toyD.classList.add("none")
            descriptionD.innerText = "Попробуй в следующий раз!"
            takeD.innerText = "закрыть"
            break;
        case 1:
            ticketD.classList.add("red")
            typeD.innerText = "Промокод"
            bodyD.innerText = "HNY23-15"
            bodyD.classList.add("promo")
            toyD.classList.add("promo")
            copyD.style.display = "block"
            copyD.addEventListener('click', async function (e) {
                navigator.clipboard.writeText("HNY23-15")
            })
            descriptionD.innerText = `Скопируй и введи в @ABUZZBOT в раздел “промокоды”`
            takeD.innerText = "закрыть"
            break;
        case 2:
            ticketD.classList.add("green")
            typeD.innerText = "УРА !"
            bodyD.innerText = "ТЫ  ВЫИГРАЛ"
            bodyD.classList.add("price")
            toyD.classList.add("price")
            descriptionD.innerText = `Зайди в личный кабинет и проверь свой выигрыш`
            takeD.innerText = "забрать"
            takeD.onclick = function () {
                window.location = "/lk"
            }
            break;
    }
    /*if (data.price.win) {
        ticketD.classList.add("red")
        typeD.innerText = "УРА !"
        bodyD.innerText = "ТЫ  ВЫИГРАЛ"
        bodyD.classList.add("promo")
        toyD.classList.add("promo")
        descriptionD.innerText = `Зайди в личный кабинет и проверь свой выигрыш`
        takeD.innerText = "забрать"
        takeD.onclick = function () {
            window.location = "/lk"
        }
    } else {
        ticketD.classList.add("green")
        typeD.innerText = "УВЫ :("
        bodyD.innerText = data.price.data
        bodyD.classList.add("none")
        toyD.classList.add("none")
        descriptionD.innerText = "Попробуй в следующий раз!"
        takeD.innerText = "закрыть"
    }*/
    let modal = document.getElementById("game-modal")
    modal.style.display = "block"
}

function closeModal(id) {
    let modal = document.getElementById(id)
    modal.style.display = "none"
}

document.getElementById("take-price").addEventListener('click', function (e) {
    closeModal("game-modal")
})

document.getElementById("modal-close").addEventListener('click', function (e) {
    closeModal("game-modal")
})

function getTimeRemaining(endtime) {
    var t = Date.parse(endtime) - Date.parse(new Date());
    var seconds = Math.floor((t / 1000) % 60);
    var minutes = Math.floor((t / 1000 / 60) % 60);
    var hours = Math.floor((t / (1000 * 60 * 60)) % 24);
    var days = Math.floor(t / (1000 * 60 * 60 * 24));
    return {
        'total': t,
        'days': days,
        'hours': hours,
        'minutes': minutes,
        'seconds': seconds
    };
}

function initializeClock(id, endtime) {
    var button = document.getElementById("start")
    button.classList.add("disable");
    var button_mobile = document.getElementById("start-mobile")
    button_mobile.classList.add("disable");
    var clock = document.getElementById(id);
    var timeinterval = setInterval(function () {
        var t = getTimeRemaining(endtime);
        clock.innerHTML = `${t.hours < 10 ? "0" + t.hours : t.hours}:${t.minutes < 10 ? "0" + t.minutes : t.minutes}:${t.seconds < 10 ? "0" + t.seconds : t.seconds}`
        if (t.total <= 0) {
            var button = document.getElementById("start")
            button.classList.remove("disable");
            var button_mobile = document.getElementById("start-mobile")
            button_mobile.classList.remove("disable");
            clearInterval(timeinterval);
        }
    }, 1000);
}

var modalI = document.getElementById("game-modal-inner");
var modal = document.getElementById("game-modal");
window.addEventListener("click", function (event) {
    if (event.target === modalI) {
        modal.style.display = "none";
    }
    if (!(event.target === menu || event.target === toggler || event.target === togglerI || event.target === togglerP)) {
        menu.classList.add("hidden")
    }
});

async function ban(tg) {
    const rawResponse = await fetch(`/admin/ban/${tg}`, {
        method: 'GET',
    });
    const content = await rawResponse.text();
    console.log(content)
}

async function unban(tg) {
    const rawResponse = await fetch(`/admin/unban/${tg}`, {
        method: 'GET',
    });
    const content = await rawResponse.text();
    console.log(content)
}

document.addEventListener('DOMContentLoaded', function () {
    startRandomItems()
    if (!baned){
        if (logined) {
            if (can) {
                var clock = document.getElementById("game-time");
                var clock_mobile = document.getElementById("game-time-mobile");
                clock.innerHTML = "00:00:00"
                clock_mobile.innerHTML = "00:00:00"
                var button = document.getElementById("start")
                var button_mobile = document.getElementById("start-mobile")
                button.classList.remove("disable");
                button_mobile.classList.remove("disable");
            } else {
                initializeClock("game-time", endTime)
                initializeClock("game-time-mobile", endTime)
            }
    }
    }
});
