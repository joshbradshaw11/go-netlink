import QtQuick 2.0
import QtQuick.Layouts 1.1
import QtQuick.Controls 1.3
import Ubuntu.Components 1.2
import GoExtensions 1.0


/*!
    \brief MainView with a Label and Button elements.
                                                */
MainView {
    objectName: "main"
    applicationName: "ibutton.mh"

    width: units.gu(100)
    height: units.gu(75)

    App {
        id: app
        status: Status {
            time: "ROFL"
            onCountChanged: if (count > 0)
                                app.readLog()
            onClearedChanged: {
                app.update()
                app.start()
            }
        }

        onStateChanged: if (state == "CONNECTED")
                            app.update()
    }

    // automatic connection handling
    // note that onDestruction is not guaranteed to be called at all
    Component.onCompleted: app.connect()
    Component.onDestruction: app.disconnect()

    Page {
        title: i18n.tr("iButton")

        ColumnLayout {

            GridLayout {
                columns: 2
                rowSpacing: units.gu(1)
                columnSpacing: units.gu(1)
                anchors {
                    margins: units.gu(2)
                }

                Label {
                    text: "Status"
                }
                Label {
                    text: app.status.missionProgress
                }

                Label {
                    text: "Started"
                }
                Label {
                    text: app.status.startedTime
                }

                Label {
                    text: "Sample count"
                }
                Label {
                    text: app.status.count
                }

                Button {
                    text: i18n.tr("Start")
                    onClicked: app.start()
                }
                Button {
                    text: i18n.tr("Stop")
                    onClicked: app.stop()
                }
            }

            Component {
                id: sampleDelegate
                Row {
                    spacing: 10
                    Text {
                        text: app.sampleTime(index)
                    }
                    Text {
                        text: app.sampleTemp(index)
                    }
                }
            }

            ListView {
                width: units.gu(90)
                height: units.gu(30)
                model: app.samples.len
                delegate: sampleDelegate
            }
        }
    }
}
