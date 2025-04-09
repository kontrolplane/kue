<p align="center">
  <h1 align="center">
    <a href="https://kontrolplane.dev">
      <img width="1500" alt="kontrolplane header" src="./assets/kontrolplane-header.svg">
    </a>
  </h1>
</p>

`Kue` is a terminal user interface (tui) application designed for managing aws sqs (simple queue service). It provides an intuitive and efficient way to interact with your sqs queues directly from the terminal. With Kue, you can easily create, delete, and manage messages within your queues, making it an essential tool for engineers who prefer working within a terminal environment.

<p align="center">
  <img width="1500" alt="kue cassette" src="./assets/cassette.gif">
</p>

## views

- `queue`: overview, details<sup>1</sup>, creation<sup>1</sup>, delete
- `message`: details<sup>1</sup>, creation<sup>1</sup>, delete<sup>1</sup>

<sup>1</sup>: work in progress

## keybindings

- `q`, `esc`, `ctrl+c`: quit/return
- `↑`, `k`: up
- `↓`, `j`: down
- `→`, `l`: right
- `←`, `h`: left
- `ctrl + d`: delete queue/message
- `ctrl + n`: create queue/message
- `?`: help
- `enter`: view
- `space`: select
- `/`: filter

## demonstration

`queue overview`
<p align="center">
  <img width="1500" alt="kue queue overview" src="./assets/pages/queue/overview.png">
</p>

`queue delete`
<p align="center">
  <img width="1500" alt="kue queue delete" src="./assets/pages/queue/delete.png">
</p>

</br>

<p align="center">
  <img width="1500" alt="kontrolplane foter" src="./assets/kontrolplane-footer.svg">
</p>