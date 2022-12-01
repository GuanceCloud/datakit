#encoding: utf-8

from datakit_framework import DataKitFramework

class Demo(DataKitFramework):
    name = 'Demo'
    interval = 10 # triggered interval seconds.

    # if your datakit ip is 127.0.0.1 and port is 9529, you won't need use this,
    # just comment it.
    # def __init__(self, **kwargs):
    #     super().__init__(ip = '127.0.0.1', port = 9529)

    # General report example.
    def run(self):
        print("Demo")
        data = [
                {
                    "measurement": "abc",
                    "tags": {
                    "t1": "b",
                    "t2": "d"
                    },
                    "fields": {
                    "f1": 123,
                    "f2": 3.4,
                    "f3": "strval"
                    },
                    # "time": 1624550216 # you don't need this
                },

                {
                    "measurement": "def",
                    "tags": {
                    "t1": "b",
                    "t2": "d"
                    },
                    "fields": {
                    "f1": 123,
                    "f2": 3.4,
                    "f3": "strval"
                    },
                    # "time": 1624550216 # you don't need this
                }
            ]

        in_data = {
            'M':data,
            'input': "datakitpy"
        }

        return self.report(in_data) # you must call self.report here

    # # KeyEvent report example.
    # def run(self):
    #     print("Demo")

    #     tags = {"tag1": "val1", "tag2": "val2"}
    #     date_range = 10
    #     status = 'info'
    #     event_id = 'event_id'
    #     title = 'title'
    #     message = 'message'
    #     kwargs = {"custom_key1":"custom_value1", "custom_key2": "custom_value2", "custom_key3": "custom_value3"}

        # Feed df_source=user event.
        # user_id="user_id"
        # return self.feed_user_event(
        #     df_user_id=user_id,
        #     tags=tags, df_date_range=date_range, df_status=status, df_event_id=event_id, df_title=title, df_message=message, **kwargs
        #     )

        # # Feed df_source=monitor event.
        # dimension_tags='{"host":"web01"}' # dimension_tags must be the String(JSON format).
        # return self.feed_monitor_event(
        #     df_dimension_tags=dimension_tags,
        #     tags=tags, df_date_range=date_range, df_status=status, df_event_id=event_id, df_title=title, df_message=message, **kwargs
        #     )

        # # Feed df_source=system event.
        # return self.feed_system_event(
        #     tags=tags, df_date_range=date_range, df_status=status, df_event_id=event_id, df_title=title, df_message=message, **kwargs
        #     )
