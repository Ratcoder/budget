module Main exposing (..)

import Browser
import Html exposing (..)
import Html.Attributes exposing (..)
import Html.Events exposing (..)
import Http exposing (..)
import Json.Decode
import Json.Encode


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
    | GetTransactions
    | GetTransactionsResponse (Result Http.Error (List Transaction))
    | GetCategories
    | GetCategoriesResponse (Result Http.Error (List Category))
    | GetAccounts
    | GetAccountsResponse (Result Http.Error (List Account))


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
                }
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
                , ul [] (List.map (\t -> li [] [ text t.description ]) user.transactions)
                , button [ onClick GetCategories ] [ text "Get Categories" ]
                , ul [] (List.map (\c -> li [] [ text c.name ]) user.categories)
                , button [ onClick GetAccounts ] [ text "Get Accounts" ]
                , ul [] (List.map (\a -> li [] [ text a.name ]) user.accounts)
                ]
            }


transactionDecoder : Json.Decode.Decoder Transaction
transactionDecoder =
    Json.Decode.map5 Transaction
        (Json.Decode.field "id" Json.Decode.int)
        (Json.Decode.field "amount" Json.Decode.int)
        (Json.Decode.field "description" Json.Decode.string)
        (Json.Decode.field "date" Json.Decode.string)
        (Json.Decode.field "category_id" Json.Decode.int)


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
