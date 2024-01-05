module Main exposing (..)

import Browser
import Html exposing (..)
import Html.Attributes exposing (..)
import Html.Events exposing (..)
import Http exposing (..)
import Json.Decode
import Json.Encode
import Platform.Cmd as Cmd


main : Program () Model Msg
main =
    Browser.document
        { init = init
        , view = view
        , update = update
        , subscriptions = \_ -> Sub.none
        }


init : () -> ( Model, Cmd Msg )
init _ =
    ( Stranger
        { usernameField = ""
        , passwordField = ""
        , responce = Unsent
        , isRegistering = False
        }
    , Cmd.none
    )


type Model
    = Stranger
        { usernameField : String
        , passwordField : String
        , responce : Response () String
        , isRegistering : Bool
        }
    | User
        { transactions : List Transaction
        , accounts : List Account
        , categories : List Category
        , dragedTransaction : Maybe Transaction
        }


type alias Transaction =
    { id : Int
    , amount : Int
    , description : String
    , date : String
    , categoryId : Int
    }


type alias Account =
    { id : Int
    , name : String
    , balance : Int
    }


type alias Category =
    { id : Int
    , name : String
    , available : Int
    , budgeted : Int
    }


type Response success failure
    = Unsent
    | Loading
    | Success success
    | Failure failure


type Msg
    = ChangeUsername String
    | ChangePassword String
    | LoginSubmit
    | LoginResponse (Result Http.Error String)
    | DragEnterTransaction Transaction
    | DropTransaction Int
    | GetTransactions
    | GetTransactionsResponse (Result Http.Error (List Transaction))
    | GetCategories
    | GetCategoriesResponse (Result Http.Error (List Category))
    | GetAccounts
    | GetAccountsResponse (Result Http.Error (List Account))
    | NoOp


update : Msg -> Model -> ( Model, Cmd Msg )
update msg model =
    case ( msg, model ) of
        ( ChangeUsername username, Stranger stranger ) ->
            ( Stranger { stranger | usernameField = username }, Cmd.none )

        ( ChangePassword password, Stranger stranger ) ->
            ( Stranger { stranger | passwordField = password }, Cmd.none )

        ( LoginSubmit, Stranger stranger ) ->
            ( Stranger { stranger | responce = Loading }
            , Http.post
                { url = "/api/login"
                , expect = Http.expectString LoginResponse
                , body =
                    Http.jsonBody <|
                        Json.Encode.object
                            [ ( "username", Json.Encode.string stranger.usernameField )
                            , ( "password", Json.Encode.string stranger.passwordField )
                            ]
                }
            )

        ( LoginResponse (Ok _), Stranger _ ) ->
            ( User
                { transactions = []
                , accounts = []
                , categories = []
                , dragedTransaction = Nothing
                }
            , Cmd.none
            )

        ( DragEnterTransaction transaction, User user ) ->
            ( User { user | dragedTransaction = Just transaction }, Cmd.none )

        ( DropTransaction categoryId, User user ) ->
            case user.dragedTransaction of
                Just transaction ->
                    ( User
                        { user
                            | dragedTransaction = Nothing
                            , transactions =
                                List.map
                                    (\t ->
                                        if t.id == transaction.id then
                                            { t | categoryId = categoryId }

                                        else
                                            t
                                    )
                                    user.transactions
                        }
                    , patch
                        { url = "/api/transactions"
                        , body =
                            Http.jsonBody <|
                                encodeTransaction
                                    { transaction | categoryId = categoryId }
                        , expect = Http.expectString (\_ -> NoOp)
                        }
                    )

                Nothing ->
                    ( model
                    , Cmd.none
                    )

        ( GetTransactions, User _ ) ->
            ( model
            , Http.get
                { url = "/api/transactions"
                , expect = Http.expectJson GetTransactionsResponse (Json.Decode.list transactionDecoder)
                }
            )

        ( GetTransactionsResponse (Ok transactions), User user ) ->
            ( User { user | transactions = transactions }
            , Cmd.none
            )

        ( GetCategories, User _ ) ->
            ( model
            , Http.get
                { url = "/api/categories"
                , expect = Http.expectJson GetCategoriesResponse (Json.Decode.list categoryDecoder)
                }
            )

        ( GetCategoriesResponse (Ok categories), User user ) ->
            ( User { user | categories = categories }
            , Cmd.none
            )

        ( GetAccounts, User _ ) ->
            ( model
            , Http.get
                { url = "/api/accounts"
                , expect = Http.expectJson GetAccountsResponse (Json.Decode.list accountDecoder)
                }
            )

        ( GetAccountsResponse (Ok accounts), User user ) ->
            ( User { user | accounts = accounts }
            , Cmd.none
            )

        _ ->
            ( model, Cmd.none )


view : Model -> Browser.Document Msg
view model =
    case model of
        Stranger stranger ->
            { title = "Budget"
            , body =
                [ h2 [] [ text "Budget" ]
                , div []
                    [ button [ onClick LoginSubmit ] [ text "Login" ]
                    , input [ type_ "text", value stranger.usernameField, onInput ChangeUsername ] []
                    , input [ type_ "password", value stranger.passwordField, onInput ChangePassword ] []
                    , p []
                        [ case stranger.responce of
                            Unsent ->
                                text ""

                            Loading ->
                                text "Loading..."

                            Success _ ->
                                text "Success!"

                            Failure err ->
                                text <| "Failure: " ++ err
                        ]
                    ]
                ]
            }

        User user ->
            { title = "Budget"
            , body =
                [ h2 [] [ text "Dashboard" ]
                , button [ onClick GetTransactions ] [ text "Get Transactions" ]
                , button [ onClick GetCategories ] [ text "Get Categories" ]
                , button [ onClick GetAccounts ] [ text "Get Accounts" ]
                , h2 [] [ text "Uncategorized Transactions:" ]
                , ul [] <| List.map (\t -> li [] [ viewTransaction t ]) <| List.filter (\t -> t.categoryId == 0) <| List.filter (\t -> t.date >= "2023-12-01") user.transactions
                , h2 [] [ text "Categories:" ]
                , ul [] <|
                    List.map
                        (\c ->
                            li [ preventDefaultOn "drop" (Json.Decode.succeed ( DropTransaction c.id, True )), preventDefaultOn "dragover" (Json.Decode.succeed ( NoOp, True )) ]
                                [ div [ style "display" "flex", style "width" "500px", style "gap" "3ch" ]
                                    [ div [ style "flex" "1" ] [ text c.name ]
                                    , div [ style "text-align" "right", style "width" "15ch" ] [ text <| String.fromInt c.available ]
                                    , div [ style "text-align" "right", style "width" "15ch" ] [ text <| String.fromInt c.budgeted ]
                                    ]
                                , ul [] <| List.map (\t -> li [] [ viewTransaction t ]) <| List.filter (\t -> t.categoryId == c.id) user.transactions
                                ]
                        )
                        user.categories
                ]
            }


viewTransaction : Transaction -> Html Msg
viewTransaction transaction =
    div [ draggable "true", on "dragstart" (Json.Decode.succeed (DragEnterTransaction transaction)) ] [ text <| transaction.description ++ " - " ++ String.fromInt transaction.amount ]


transactionDecoder : Json.Decode.Decoder Transaction
transactionDecoder =
    Json.Decode.map5 Transaction
        (Json.Decode.field "id" Json.Decode.int)
        (Json.Decode.field "amount" Json.Decode.int)
        (Json.Decode.field "description" Json.Decode.string)
        (Json.Decode.field "date" Json.Decode.string)
        (Json.Decode.field "category_id" Json.Decode.int)


encodeTransaction : Transaction -> Json.Encode.Value
encodeTransaction transaction =
    Json.Encode.object
        [ ( "id", Json.Encode.int transaction.id )
        , ( "amount", Json.Encode.int transaction.amount )
        , ( "description", Json.Encode.string transaction.description )
        , ( "date", Json.Encode.string transaction.date )
        , ( "category_id", Json.Encode.int transaction.categoryId )
        ]


categoryDecoder : Json.Decode.Decoder Category
categoryDecoder =
    Json.Decode.map4 Category
        (Json.Decode.field "id" Json.Decode.int)
        (Json.Decode.field "name" Json.Decode.string)
        (Json.Decode.field "available" Json.Decode.int)
        (Json.Decode.field "budgeted" Json.Decode.int)


accountDecoder : Json.Decode.Decoder Account
accountDecoder =
    Json.Decode.map3 Account
        (Json.Decode.field "id" Json.Decode.int)
        (Json.Decode.field "name" Json.Decode.string)
        (Json.Decode.field "balance" Json.Decode.int)


patch :
    { url : String
    , body : Http.Body
    , expect : Http.Expect msg
    }
    -> Cmd msg
patch p =
    Http.request
        { method = "PATCH"
        , headers = []
        , url = p.url
        , body = p.body
        , expect = p.expect
        , timeout = Nothing
        , tracker = Nothing
        }
